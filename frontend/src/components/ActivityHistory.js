import React, { useState, useEffect, useCallback } from 'react';
import { useAuth } from '../context/AuthContext';
import api from '../services/api';

const ActivityHistory = () => {
  const { user, isAuthenticated } = useAuth();
  const [activities, setActivities] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [filter, setFilter] = useState('all'); // all, SONG_PLAYED, RATING_GIVEN, etc.

  const loadActivities = useCallback(async () => {
    console.log('loadActivities called with filter:', filter, 'user:', user);
    setLoading(true);
    setError('');
    try {
      const type = filter === 'all' ? null : filter;
      console.log('Loading activities with filter:', filter, 'type:', type);
      const data = await api.getUserActivities(100, type, user?.id);
      console.log('Received activities data:', data);
      console.log('Activities array length:', Array.isArray(data) ? data.length : 'not an array');
      setActivities(Array.isArray(data) ? data : []);
    } catch (err) {
      console.error('Error loading activities:', err);
      console.error('Error details:', err.message, err.stack);
      setError(err.message || 'Greška pri učitavanju istorije aktivnosti');
    } finally {
      setLoading(false);
    }
  }, [filter, user]);

  useEffect(() => {
    console.log('ActivityHistory useEffect triggered:', { isAuthenticated, user: user?.id, filter });
    if (isAuthenticated && user) {
      console.log('User is authenticated, loading activities...');
      loadActivities();
    } else {
      console.log('User not authenticated or user object missing:', { isAuthenticated, user });
      setLoading(false);
    }
  }, [isAuthenticated, user, filter, loadActivities]);

  const getActivityIcon = (type) => {
    switch (type) {
      case 'SONG_PLAYED':
        return '🎵';
      case 'RATING_GIVEN':
        return '⭐';
      case 'GENRE_SUBSCRIBED':
        return '➕';
      case 'GENRE_UNSUBSCRIBED':
        return '➖';
      case 'ARTIST_SUBSCRIBED':
        return '➕';
      case 'ARTIST_UNSUBSCRIBED':
        return '➖';
      default:
        return '📝';
    }
  };

  const getActivityText = (activity) => {
    switch (activity.type) {
      case 'SONG_PLAYED':
        return `Slušali ste pesmu "${activity.songName || activity.songId}"`;
      case 'RATING_GIVEN':
        return `Ocenili ste pesmu "${activity.songName || activity.songId}" sa ${activity.rating} zvezdica`;
      case 'GENRE_SUBSCRIBED':
        return `Pretplatili ste se na žanr "${activity.genre}"`;
      case 'GENRE_UNSUBSCRIBED':
        return `Odpretplatili ste se od žanra "${activity.genre}"`;
      case 'ARTIST_SUBSCRIBED':
        return `Pretplatili ste se na umetnika "${activity.artistName || activity.artistId}"`;
      case 'ARTIST_UNSUBSCRIBED':
        return `Odpretplatili ste se od umetnika "${activity.artistName || activity.artistId}"`;
      default:
        return `Aktivnost: ${activity.type}`;
    }
  };

  const formatDate = (timestamp) => {
    if (!timestamp) return '';
    const date = new Date(timestamp);
    return date.toLocaleString('sr-RS', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  if (!isAuthenticated) {
    return (
      <div style={{ padding: '40px', textAlign: 'center' }}>
        <p>Morate biti prijavljeni da biste videli istoriju aktivnosti.</p>
      </div>
    );
  }

  return (
    <div style={{ padding: '40px', maxWidth: '1200px', margin: '0 auto' }}>
      <h1 style={{ marginBottom: '30px', fontSize: '32px', fontWeight: '600' }}>
        📊 Istorija Aktivnosti
      </h1>

      {/* Filter */}
      <div style={{ marginBottom: '30px', display: 'flex', gap: '10px', flexWrap: 'wrap' }}>
        <button
          onClick={() => setFilter('all')}
          style={{
            padding: '10px 20px',
            borderRadius: '8px',
            border: 'none',
            background: filter === 'all' ? 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)' : '#f0f0f0',
            color: filter === 'all' ? 'white' : '#333',
            cursor: 'pointer',
            fontWeight: '500',
            transition: 'all 0.3s ease',
          }}
        >
          Sve
        </button>
        <button
          onClick={() => setFilter('SONG_PLAYED')}
          style={{
            padding: '10px 20px',
            borderRadius: '8px',
            border: 'none',
            background: filter === 'SONG_PLAYED' ? 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)' : '#f0f0f0',
            color: filter === 'SONG_PLAYED' ? 'white' : '#333',
            cursor: 'pointer',
            fontWeight: '500',
            transition: 'all 0.3s ease',
          }}
        >
          🎵 Slušanje
        </button>
        <button
          onClick={() => setFilter('RATING_GIVEN')}
          style={{
            padding: '10px 20px',
            borderRadius: '8px',
            border: 'none',
            background: filter === 'RATING_GIVEN' ? 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)' : '#f0f0f0',
            color: filter === 'RATING_GIVEN' ? 'white' : '#333',
            cursor: 'pointer',
            fontWeight: '500',
            transition: 'all 0.3s ease',
          }}
        >
          ⭐ Ocene
        </button>
        <button
          onClick={() => setFilter('GENRE_SUBSCRIBED')}
          style={{
            padding: '10px 20px',
            borderRadius: '8px',
            border: 'none',
            background: filter === 'GENRE_SUBSCRIBED' ? 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)' : '#f0f0f0',
            color: filter === 'GENRE_SUBSCRIBED' ? 'white' : '#333',
            cursor: 'pointer',
            fontWeight: '500',
            transition: 'all 0.3s ease',
          }}
        >
          ➕ Pretplate na žanrove
        </button>
        <button
          onClick={() => setFilter('ARTIST_SUBSCRIBED')}
          style={{
            padding: '10px 20px',
            borderRadius: '8px',
            border: 'none',
            background: filter === 'ARTIST_SUBSCRIBED' ? 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)' : '#f0f0f0',
            color: filter === 'ARTIST_SUBSCRIBED' ? 'white' : '#333',
            cursor: 'pointer',
            fontWeight: '500',
            transition: 'all 0.3s ease',
          }}
        >
          ➕ Pretplate na umetnike
        </button>
      </div>

      {error && (
        <div style={{
          background: 'linear-gradient(135deg, rgba(255, 193, 7, 0.1) 0%, rgba(255, 152, 0, 0.1) 100%)',
          border: '2px solid rgba(255, 193, 7, 0.3)',
          borderRadius: '12px',
          padding: '20px',
          marginBottom: '20px',
          textAlign: 'center',
        }}>
          <p style={{ color: '#666', fontSize: '16px' }}>{error}</p>
        </div>
      )}

      {loading ? (
        <div style={{ textAlign: 'center', padding: '40px' }}>
          <p>Učitavanje aktivnosti...</p>
        </div>
      ) : activities.length === 0 ? (
        <div style={{
          background: 'linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%)',
          border: '2px solid rgba(102, 126, 234, 0.3)',
          borderRadius: '12px',
          padding: '40px',
          textAlign: 'center',
        }}>
          <p style={{ color: '#666', fontSize: '18px' }}>
            {filter === 'all' 
              ? 'Nemate još aktivnosti. Počnite da slušate pesme, ocenjujete ih i pretplaćujete se na žanrove i umetnike!'
              : 'Nemate aktivnosti ovog tipa.'}
          </p>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
          {activities.map((activity) => (
            <div
              key={activity.id}
              style={{
                background: 'white',
                borderRadius: '12px',
                padding: '20px',
                boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
                display: 'flex',
                alignItems: 'center',
                gap: '20px',
                transition: 'all 0.3s ease',
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.transform = 'translateY(-2px)';
                e.currentTarget.style.boxShadow = '0 4px 12px rgba(0,0,0,0.15)';
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.transform = 'translateY(0)';
                e.currentTarget.style.boxShadow = '0 2px 8px rgba(0,0,0,0.1)';
              }}
            >
              <div style={{ fontSize: '32px' }}>
                {getActivityIcon(activity.type)}
              </div>
              <div style={{ flex: 1 }}>
                <p style={{ fontSize: '16px', fontWeight: '500', marginBottom: '4px', color: '#333' }}>
                  {getActivityText(activity)}
                </p>
                <p style={{ fontSize: '14px', color: '#666' }}>
                  {formatDate(activity.timestamp)}
                </p>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default ActivityHistory;
