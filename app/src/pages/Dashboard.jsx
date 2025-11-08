import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import ImageUpload from '../components/ImageUpload';
import ImageGallery from '../components/ImageGallery';
import { jwtDecode } from 'jwt-decode';

const Dashboard = () => {
  const [activeTab, setActiveTab] = useState('gallery');
  const [userRole, setUserRole] = useState('');
  const [userName, setUserName] = useState('');
  const { logout } = useAuth();

  // ✅ Decode token and extract user info
  useEffect(() => {
    const token = localStorage.getItem('token');
    if (token) {
      try {
        const decoded = jwtDecode(token);
        setUserRole(decoded?.role || '');
        setUserName(decoded?.first_name || decoded?.email || 'User');
      } catch (err) {
        console.error('Error decoding token:', err);
      }
    }
  }, []);

  const handleLogout = () => {
    logout();
  };

  return (
    <div className="dashboard">
      <header className="dashboard-header">
        <div className="dashboard-info">
          <h1>Image Store</h1>
          {userName && <p className="welcome-text">Welcome, {userName}!</p>}
          {userRole && <span className="user-role-badge">{userRole}</span>}
        </div>
        <nav className="dashboard-nav">
          <button
            className={activeTab === 'gallery' ? 'active' : ''}
            onClick={() => setActiveTab('gallery')}
          >
            Gallery
          </button>
          
          {/* ✅ Only show Upload tab for admin users */}
          {userRole === 'admin' && (
            <button
              className={activeTab === 'upload' ? 'active' : ''}
              onClick={() => setActiveTab('upload')}
            >
              Upload
            </button>
          )}
          
          <button onClick={handleLogout} className="logout-btn">
            Logout
          </button>
        </nav>
      </header>
      
      <main className="dashboard-content">
        {activeTab === 'gallery' && <ImageGallery />}
        {activeTab === 'upload' && userRole === 'admin' && (
          <ImageUpload onUploadSuccess={() => setActiveTab('gallery')} />
        )}
      </main>
    </div>
  );
};

export default Dashboard;