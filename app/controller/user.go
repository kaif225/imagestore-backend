package controller

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"ginmongo/database"
	"ginmongo/models"
	"ginmongo/utils"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func RegisterUser(c *gin.Context) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var user models.User
	err := c.ShouldBind(&user)

	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid Request Body"})
		return
	}

	validate := validator.New()

	if err = validate.Struct(user); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation Failed", "details": err.Error()})
		return
	}
	HashPas, err := utils.HashPass(user.Password)

	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Error Hashing Password"})
		return
	}

	collection := database.Client.Database("imagestore").Collection("users")
	count, err := collection.CountDocuments(ctx, bson.M{"email": user.Email})

	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing user"})
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User already exist"})
		return
	}
	user.UserID = bson.NewObjectID().Hex()
	user.Password = HashPas
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Role = "user"

	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Error Adding user"})
		return
	}

	c.IndentedJSON(http.StatusOK, result)
}

func Login(c *gin.Context) {

	var userLogin models.UserLogin

	err := c.ShouldBind(&userLogin)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid Request Body"})
		return
	}
	validate := validator.New()
	err = validate.Struct(userLogin)
	if err != nil {
		log.Println(err)
		return
	}
	if userLogin.Email == "" || userLogin.Password == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Both are required"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := database.Client.Database("imagestore").Collection("users")

	userExist := &models.User{}

	err = collection.FindOne(ctx, bson.M{"email": userLogin.Email}).Decode(userExist)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}
	err = utils.ComparePass(userLogin.Password, userExist.Password)
	if err != nil {
		log.Println(err)
		return
	}

	token, err := utils.SignedToken(userLogin.Email, userExist.FirstName, userExist.LastName,
		userExist.Role)

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "Bearer",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(1 * time.Hour),
		Secure:   false,
		HttpOnly: true,
		//SameSite: http.SameSiteStrictMode,
		SameSite: http.SameSiteLaxMode,
	})

	response := struct {
		Status string `json:"status"`
		Token  string `json:"token"`
	}{
		Status: "Login Successfull",
		Token:  token,
	}
	c.IndentedJSON(http.StatusOK, response)
}

func Logout(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "Bearer",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Second),
		MaxAge:   -1,
		Secure:   false,
		HttpOnly: true,
		//SameSite: http.SameSiteStrictMode,
		SameSite: http.SameSiteLaxMode,
	})

	response := struct {
		Status string `json:"status"`
	}{
		Status: "Logout Successfull",
	}
	c.IndentedJSON(http.StatusOK, response)
}

func UpdatePassword(c *gin.Context) {
	userId := c.Param("user_id")

	if userId == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Id is missing"})
		return
	}
	var updatePass models.PasswordUpdate
	if err := c.ShouldBind(&updatePass); err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid Payload"})
		return
	}

	validate := validator.New()

	if err := validate.Struct(updatePass); err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := database.Client.Database("imagestore").Collection("users")

	user := &models.User{}

	err := collection.FindOne(ctx, bson.M{"user_id": userId}).Decode(user)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Error getting data"})
		return
	}

	if updatePass.CurrentPassword == "" || user.Password == "" {
		c.IndentedJSON(http.StatusOK, gin.H{"error": "Both are required"})
		return
	}
	err = utils.ComparePass(updatePass.CurrentPassword, user.Password)

	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusOK, gin.H{"error": "Incorrect Password"})
		return
	}
	newHashPass, err := utils.HashPass(updatePass.NewPassword)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusOK, gin.H{"error": "Incorrect encrypting password"})
		return
	}

	user.Password = newHashPass
	user.UpdatedAt = time.Now()

	_, err = collection.UpdateOne(
		ctx,
		bson.M{"user_id": user.UserID},
		bson.M{"$set": bson.M{
			"password":   user.Password,
			"updated_at": user.UpdatedAt,
		}},
	)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

func ForgetPassword(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	type emailRequest struct {
		Email string `json:"email" validate:"required,email"`
	}

	var req emailRequest

	if err := c.ShouldBind(&req); err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid Request Body"})
		return
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Email == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}

	user := &models.User{}
	collection := database.Client.Database("imagestore").Collection("users")
	err := collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(user)

	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	duration, err := strconv.Atoi(os.Getenv("RESET_TOKEN_EXPIRY"))
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Error converting token expiry"})
		return
	}
	mins := time.Duration(duration)

	expiry := time.Now().Add(mins * time.Minute)

	tokenByte := make([]byte, 16)
	_, err = rand.Read(tokenByte)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Error creating byte value"})
		return
	}

	token := hex.EncodeToString(tokenByte)
	hashedToken := sha256.Sum256(tokenByte)

	hashTokenStr := hex.EncodeToString(hashedToken[:])

	_, err = collection.UpdateOne(
		ctx,
		bson.M{"email": req.Email},
		bson.M{"$set": bson.M{
			"password_reset_token":   hashTokenStr,
			"password_token_expired": expiry,
		}},
	)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Error updating token"})
		return
	}
	resetURL := fmt.Sprintf("http://localhost:8007/users/resetpassword/reset/%s", token)
	message := fmt.Sprintf("Forgot your password ? Reset using the following link: \n%s\n If you didn't request the password reset ignore this message", resetURL)

	smtpUser := os.Getenv("SMTP_USER")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpPass := os.Getenv("SMTP_PASSWORD")

	smtpClient, err := smtp.Dial(smtpHost + ":" + smtpPort)

	if err != nil {
		log.Println(err)
		return
	}

	err = smtpClient.StartTLS(&tls.Config{ServerName: smtpHost})
	if err != nil {
		log.Println(err)
		return
	}

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	if err = smtpClient.Auth(auth); err != nil {
		log.Fatal(err)
		return
	}
	// Set the sender and recipient first
	if err := smtpClient.Mail(smtpUser); err != nil {
		log.Fatal(err)
		return
	}
	if err := smtpClient.Rcpt(req.Email); err != nil {
		log.Fatal(err)
		return
	}

	// Send the email body.
	wc, err := smtpClient.Data()
	if err != nil {
		log.Fatal(err)
		return
	}
	emailBody := fmt.Sprintf("To: %s\r\nSubject: Password Reset Request\r\n\r\n%s", req.Email, message)
	_, err = fmt.Fprintf(wc, emailBody)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = wc.Close()
	if err != nil {
		log.Fatal(err)
		return
	}

	// Send the QUIT command and close the connection.
	err = smtpClient.Quit()
	if err != nil {
		log.Fatal(err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "mail has been send"})
}

func ResetPassword(c *gin.Context) {
	token := c.Param("resetcode")
	var resetPass struct {
		NewPassword     string `json:"new_password" validate:"required,max=16,min=6"`
		ConfirmPassword string `json:"confirm_password" validate:"required,max=16,min=6"`
	}

	err := c.ShouldBindJSON(&resetPass)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid RequestBody"})
		return
	}
	validate := validator.New()

	err = validate.Struct(resetPass)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if resetPass.NewPassword == "" || resetPass.ConfirmPassword == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Both are required"})
		return
	}
	if resetPass.NewPassword != resetPass.ConfirmPassword {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Password Do not match"})
		return
	}

	user := &models.User{}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	bytes, err := hex.DecodeString(token)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": ""})
		return
	}
	hashToken := sha256.Sum256(bytes)
	hashTokenStr := hex.EncodeToString(hashToken[:])
	collection := database.Client.Database("imagestore").Collection("users")

	err = collection.FindOne(ctx, bson.M{
		"password_reset_token":   hashTokenStr,
		"password_token_expired": bson.M{"$gt": time.Now()},
	}).Decode(user)

	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid token or expiry"})
		return
	}

	HashPass, err := utils.HashPass(resetPass.NewPassword)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Error hashing password"})
		return
	}
	user.Password = HashPass
	user.UpdatedAt = time.Now()

	_, err = collection.UpdateOne(
		ctx,
		bson.M{"user_id": user.UserID},
		bson.M{"$set": bson.M{
			"password":               user.Password,
			"password_reset_token":   "",
			"password_token_expired": time.Time{},
			"updated_at":             user.UpdatedAt,
		}},
	)

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Password Updated Successfully"})
}
