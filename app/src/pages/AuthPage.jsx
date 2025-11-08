import React, { useState } from 'react';
import Login from '../components/Login';
import Register from '../components/Register';
import ForgotPassword from '../components/ForgotPassword';

const AuthPage = () => {
  const [mode, setMode] = useState('login');

  const renderContent = () => {
    switch (mode) {
      case 'login':
        return <Login onToggleMode={() => setMode('register')} />;
      case 'register':
        return <Register onToggleMode={() => setMode('login')} />;
      case 'forgot':
        return <ForgotPassword onBackToLogin={() => setMode('login')} />;
      default:
        return <Login onToggleMode={() => setMode('register')} />;
    }
  };

  return (
    <div className="auth-page">
      <div className="auth-container">
        {renderContent()}
        {mode === 'login' && (
          <button 
            type="button" 
            onClick={() => setMode('forgot')}
            className="forgot-password-link link-button"
          >
            Forgot your password?
          </button>
        )}
      </div>
    </div>
  );
};

export default AuthPage;