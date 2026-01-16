import React, { createContext, useState, useContext, useEffect } from 'react';
import { setEncryptedItem, getEncryptedItem, removeEncryptedItem } from '../utils/encryption';

const AuthContext = createContext();

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
};

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Check if user is logged in from encrypted localStorage
    const token = localStorage.getItem('token'); // Token is stored as-is for API requests
    const userData = getEncryptedItem('user'); // User data is encrypted
    
    if (token && userData) {
      try {
        setUser(userData);
      } catch (e) {
        localStorage.removeItem('token');
        removeEncryptedItem('user');
      }
    }
    setLoading(false);
  }, []);

  const login = (userData, token) => {
    // Store token as-is (JWT tokens are already encoded)
    localStorage.setItem('token', token);
    // Encrypt user data for integrity and basic protection
    setEncryptedItem('user', userData);
    setUser(userData);
  };

  const logout = () => {
    localStorage.removeItem('token');
    removeEncryptedItem('user');
    setUser(null);
  };

  const isAdmin = () => {
    return user && user.role === 'ADMIN';
  };

  const value = {
    user,
    login,
    logout,
    isAdmin,
    isAuthenticated: !!user,
    loading,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};
