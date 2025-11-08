import React, { useState, useEffect } from 'react';
import { imageAPI, categoryAPI } from '../services/api';
import { jwtDecode } from 'jwt-decode';

const ImageUpload = ({ onUploadSuccess }) => {
  const [categories, setCategories] = useState([]);
  const [category, setCategory] = useState('');
  const [file, setFile] = useState(null);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState('');
  const [userRole, setUserRole] = useState('');

  // ✅ Decode token and extract user role
  useEffect(() => {
    const token = localStorage.getItem('token');
    if (token) {
      try {
        const decoded = jwtDecode(token);
        if (decoded?.role) {
          setUserRole(decoded.role);
        }
      } catch (err) {
        console.error('Error decoding token:', err);
      }
    }
  }, []);

  // ✅ Fetch categories for admin
  useEffect(() => {
    if (userRole === 'admin') {
      fetchCategories();
    }
  }, [userRole]);

  const fetchCategories = async () => {
    try {
      const response = await categoryAPI.getAll();
      setCategories(response.data.Category || []);
    } catch (error) {
      console.error('Error fetching categories:', error);
    }
  };

  const handleFileChange = (e) => {
    const selectedFile = e.target.files[0];
    if (selectedFile) {
      if (!selectedFile.type.startsWith('image/')) {
        setMessage('Please select an image file');
        return;
      }
      setFile(selectedFile);
      setMessage('');
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!category || !file) {
      setMessage('Please select both category and image');
      return;
    }

    setLoading(true);
    setMessage('');

    const formData = new FormData();
    formData.append('image', file);

    try {
      await imageAPI.upload(category, formData);
      setMessage('Image uploaded successfully!');
      setCategory('');
      setFile(null);
      document.getElementById('file-input').value = '';
      onUploadSuccess?.();
    } catch (error) {
      const errorMsg = error.response?.data?.error || 'Upload failed';
      setMessage(errorMsg);
      
      // If unauthorized, show specific message
      if (error.response?.status === 401) {
        setMessage('You are not authorized to upload images. Admin access required.');
      }
    } finally {
      setLoading(false);
    }
  };

  // ✅ Only render for admin users
  if (userRole !== 'admin') {
    return (
      <div className="image-upload">
        <div className="error-message">
          Access Denied: Only administrators can upload images.
        </div>
      </div>
    );
  }

  return (
    <div className="image-upload">
      <h3>Upload Image</h3>
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label>Category:</label>
          {categories.length > 0 ? (
            <select
              value={category}
              onChange={(e) => setCategory(e.target.value)}
              required
            >
              <option value="">Select a category</option>
              {categories.map((cat, index) => (
                <option key={index} value={cat.category}>
                  {cat.category}
                </option>
              ))}
            </select>
          ) : (
            <input
              type="text"
              value={category}
              onChange={(e) => setCategory(e.target.value)}
              placeholder="Enter category name"
              required
            />
          )}
        </div>
        <div className="form-group">
          <label>Image:</label>
          <input
            id="file-input"
            type="file"
            accept="image/*"
            onChange={handleFileChange}
            required
          />
          {file && (
            <p className="file-info">Selected: {file.name}</p>
          )}
        </div>
        {message && (
          <div className={`message ${message.includes('success') ? 'success-message' : 'error-message'}`}>
            {message}
          </div>
        )}
        <button type="submit" disabled={loading}>
          {loading ? 'Uploading...' : 'Upload Image'}
        </button>
      </form>
    </div>
  );
};

export default ImageUpload;