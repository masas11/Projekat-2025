import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import api from '../services/api';

const Notifications = () => {
  const { user } = useAuth();
  const [notifications, setNotifications] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    if (user && user.id) {
      loadNotifications();
    } else {
      setError('Morate biti prijavljeni da biste videli notifikacije');
      setLoading(false);
    }
  }, [user]);

  const loadNotifications = async () => {
    try {
      // API Gateway će automatski koristiti userId iz JWT tokena
      // Ne treba više slati userId u query parametru
      const data = await api.getNotifications();
      setNotifications(Array.isArray(data) ? data : []);
    } catch (err) {
      setError(err.message || 'Greška pri učitavanju notifikacija');
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return <div className="container">Učitavanje...</div>;
  }

  return (
    <div className="container">
      <div className="card">
        <h2>Notifikacije</h2>
        {error && <div className="error">{error}</div>}
        
        {notifications.length === 0 ? (
          <p>Nema notifikacija.</p>
        ) : (
          <div style={{ marginTop: '20px' }}>
            {notifications.map((notification) => (
              <div key={notification.id} className="list-item">
                <h3>{notification.title || 'Notifikacija'}</h3>
                {notification.message && <p>{notification.message}</p>}
                {notification.createdAt && (
                  <p style={{ fontSize: '12px', color: '#666', marginTop: '5px' }}>
                    {new Date(notification.createdAt).toLocaleString()}
                  </p>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default Notifications;
