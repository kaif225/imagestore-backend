import axios from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

const api = axios.create({
  baseURL: API_BASE_URL,
  withCredentials: true,
});

// ✅ Request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// ✅ Response interceptor to handle token expiry and redirect
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      window.location.href = '/auth';
    }
    return Promise.reject(error);
  }
);

// ✅ Auth APIs
export const authAPI = {
  register: (userData) => api.post('/registration', userData),
  login: (credentials) => api.post('/login', credentials),
  logout: () => api.post('/logout'),
  
  // ✅ Forgot / Reset / Update password
  forgotPassword: (data) => api.post('/forgetpassword', data),
  resetPassword: (resetCode, passwords) =>
    api.post(`/users/resetpassword/reset/${resetCode}`, passwords),
  updatePassword: (userId, passwordData) =>
    api.post(`/users/${userId}`, passwordData),
};

// ✅ Image APIs
export const imageAPI = {
  upload: (category, formData) =>
    api.post(`/upload/${category}`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    }),
  
  // Fixed: Added page parameter support
  getByCategory: (category, page = 1, limit = 6) => 
    api.get(`/images/${category}`, { params: { page, limit } }),
  
  getAll: (page = 1, limit = 6) =>
    api.get('/images', { params: { page, limit } }),
  
  // Fixed: Added page parameter support
  search: (name, page = 1, limit = 6) => 
    api.get('/images/search', { params: { name, page, limit } }),
};

// ✅ Category APIs (Admin only)
export const categoryAPI = {
  getAll: () => api.get('/category'),
  create: (data) => api.post('/category', data),
};

export default api;