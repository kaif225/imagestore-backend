import React, { useState } from 'react';
import { useAuth } from '../context/AuthContext';
import { validateEmail } from '../utils/helpers';

const ForgotPassword = ({ onBackToLogin }) => {
  const [email, setEmail] = useState('');
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(false);
  const { forgotPassword } = useAuth();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setMessage('');

    if (!validateEmail(email)) {
      setMessage('Please enter a valid email address');
      setLoading(false);
      return;
    }

    const result = await forgotPassword(email);
    
    if (result.success) {
      setMessage('Password reset instructions have been sent to your email');
    } else {
      setMessage(result.error);
    }
    
    setLoading(false);
  };

  return (
    <div className="auth-form">
      <h2>Forgot Password</h2>
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label>Email:</label>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </div>
        {message && (
          <div className={message.includes('sent') ? 'success-message' : 'error-message'}>
            {message}
          </div>
        )}
        <button type="submit" disabled={loading}>
          {loading ? 'Sending...' : 'Send Reset Instructions'}
        </button>
      </form>
      <button type="button" onClick={onBackToLogin} className="link-button">
        Back to Login
      </button>
    </div>
  );
};

export default ForgotPassword;