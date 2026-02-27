import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import api from '../services/api';

const Analytics = () => {
  const { user, isAuthenticated, isAdmin } = useAuth();
  const navigate = useNavigate();
  const [analytics, setAnalytics] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login');
      return;
    }
    if (isAdmin()) {
      navigate('/');
      return;
    }
    loadAnalytics();
  }, [isAuthenticated, navigate, user]);

  const loadAnalytics = async () => {
    if (!user) return;
    try {
      setLoading(true);
      setError('');
      const data = await api.getUserAnalytics(user.id);
      setAnalytics(data);
    } catch (err) {
      console.error('Greška pri učitavanju analitika:', err);
      setError(err.message || 'Greška pri učitavanju analitika');
    } finally {
      setLoading(false);
    }
  };

  if (!isAuthenticated || isAdmin()) {
    return null;
  }

  return (
    <div className="container" style={{ paddingTop: '40px', paddingBottom: '40px' }}>
      <div className="card" style={{ marginBottom: '30px' }}>
        <h2 style={{ marginBottom: '10px', color: '#333' }}>📊 Analitike</h2>
        <p style={{ color: '#666', margin: 0 }}>
          Pregled vaših statistika i aktivnosti na platformi
        </p>
      </div>

      {error && (
        <div className="error" style={{ marginBottom: '20px' }}>
          {error}
          <button
            onClick={loadAnalytics}
            style={{
              marginLeft: '10px',
              padding: '5px 10px',
              background: '#667eea',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer'
            }}
          >
            Pokušaj ponovo
          </button>
        </div>
      )}

      {loading ? (
        <div className="card">
          <p style={{ textAlign: 'center', padding: '40px' }}>Učitavanje analitika...</p>
        </div>
      ) : analytics ? (
        <>
          {/* Overview Cards */}
          <div style={{ 
            display: 'grid', 
            gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))', 
            gap: '20px',
            marginBottom: '30px'
          }}>
            <div style={{ 
              padding: '25px', 
              background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
              color: 'white',
              borderRadius: '12px',
              boxShadow: '0 4px 15px rgba(102, 126, 234, 0.3)'
            }}>
              <h3 style={{ margin: '0 0 10px 0', fontSize: '16px', opacity: 0.9 }}>
                Ukupno odslušanih pesama
              </h3>
              <p style={{ fontSize: '3em', margin: 0, fontWeight: 'bold' }}>
                {analytics.totalSongsPlayed || 0}
              </p>
            </div>

            <div style={{ 
              padding: '25px', 
              background: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)',
              color: 'white',
              borderRadius: '12px',
              boxShadow: '0 4px 15px rgba(245, 87, 108, 0.3)'
            }}>
              <h3 style={{ margin: '0 0 10px 0', fontSize: '16px', opacity: 0.9 }}>
                Prosek ocena
              </h3>
              <p style={{ fontSize: '3em', margin: 0, fontWeight: 'bold' }}>
                {analytics.averageRating ? analytics.averageRating.toFixed(2) : '0.00'}
              </p>
            </div>

            <div style={{ 
              padding: '25px', 
              background: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)',
              color: 'white',
              borderRadius: '12px',
              boxShadow: '0 4px 15px rgba(79, 172, 254, 0.3)'
            }}>
              <h3 style={{ margin: '0 0 10px 0', fontSize: '16px', opacity: 0.9 }}>
                Pretplaćeni umetnici
              </h3>
              <p style={{ fontSize: '3em', margin: 0, fontWeight: 'bold' }}>
                {analytics.subscribedArtistsCount || 0}
              </p>
            </div>
          </div>

          {/* Songs by Genre */}
          {analytics.songsPlayedByGenre && Object.keys(analytics.songsPlayedByGenre).length > 0 && (
            <div className="card" style={{ marginBottom: '30px' }}>
              <h3 style={{ marginBottom: '20px', color: '#333' }}>
                🎵 Odslušane pesme po žanru
              </h3>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: '12px' }}>
                {Object.entries(analytics.songsPlayedByGenre)
                  .sort((a, b) => b[1] - a[1]) // Sort by count descending
                  .map(([genre, count]) => (
                    <div
                      key={genre}
                      style={{
                        padding: '12px 20px',
                        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                        color: 'white',
                        borderRadius: '25px',
                        display: 'flex',
                        alignItems: 'center',
                        gap: '10px',
                        boxShadow: '0 2px 8px rgba(102, 126, 234, 0.2)',
                        transition: 'transform 0.2s',
                        cursor: 'default'
                      }}
                      onMouseEnter={(e) => e.target.style.transform = 'scale(1.05)'}
                      onMouseLeave={(e) => e.target.style.transform = 'scale(1)'}
                    >
                      <span style={{ fontWeight: '500', fontSize: '16px' }}>{genre}</span>
                      <span style={{ 
                        fontWeight: 'bold', 
                        fontSize: '18px',
                        background: 'rgba(255, 255, 255, 0.2)',
                        padding: '4px 10px',
                        borderRadius: '15px'
                      }}>
                        {count}
                      </span>
                    </div>
                  ))}
              </div>
            </div>
          )}

          {/* Top 5 Artists */}
          {analytics.top5Artists && analytics.top5Artists.length > 0 && (
            <div className="card">
              <h3 style={{ marginBottom: '20px', color: '#333' }}>
                🎤 Top 5 umetnika koje najviše slušate
              </h3>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                {analytics.top5Artists.map((artist, index) => (
                  <div
                    key={artist.artistId}
                    style={{
                      padding: '18px',
                      background: index === 0 
                        ? 'linear-gradient(135deg, #ffd700 0%, #ffed4e 100%)'
                        : index === 1
                        ? 'linear-gradient(135deg, #c0c0c0 0%, #e8e8e8 100%)'
                        : index === 2
                        ? 'linear-gradient(135deg, #cd7f32 0%, #e6a857 100%)'
                        : '#f9f9f9',
                      borderRadius: '12px',
                      display: 'flex',
                      justifyContent: 'space-between',
                      alignItems: 'center',
                      border: index < 3 ? 'none' : '1px solid #ddd',
                      boxShadow: index < 3 ? '0 4px 12px rgba(0,0,0,0.1)' : 'none',
                      transition: 'transform 0.2s',
                      cursor: 'pointer'
                    }}
                    onMouseEnter={(e) => {
                      if (index >= 3) e.target.style.transform = 'translateX(5px)';
                    }}
                    onMouseLeave={(e) => {
                      if (index >= 3) e.target.style.transform = 'translateX(0)';
                    }}
                    onClick={() => navigate(`/artists/${artist.artistId}`)}
                  >
                    <div style={{ display: 'flex', alignItems: 'center', gap: '15px' }}>
                      <span style={{ 
                        fontSize: '1.5em', 
                        fontWeight: 'bold', 
                        color: index < 3 ? '#333' : '#667eea',
                        minWidth: '40px',
                        textAlign: 'center'
                      }}>
                        #{index + 1}
                      </span>
                      <span style={{ 
                        fontWeight: '600', 
                        fontSize: '18px',
                        color: index < 3 ? '#333' : '#333'
                      }}>
                        {artist.artistName || artist.artistId}
                      </span>
                    </div>
                    <span style={{ 
                      color: index < 3 ? '#333' : '#666',
                      fontWeight: '500',
                      fontSize: '16px'
                    }}>
                      {artist.playCount} {artist.playCount === 1 ? 'odslušanje' : 'odslušanja'}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {(!analytics.songsPlayedByGenre || Object.keys(analytics.songsPlayedByGenre).length === 0) &&
           (!analytics.top5Artists || analytics.top5Artists.length === 0) && (
            <div className="card">
              <p style={{ textAlign: 'center', color: '#666', padding: '40px' }}>
                Još nemate dovoljno aktivnosti za prikaz analitika. 
                Slušajte pesme i ocenjujte ih da biste videli svoje statistike!
              </p>
            </div>
          )}
        </>
      ) : (
        <div className="card">
          <p style={{ textAlign: 'center', color: '#666', padding: '40px' }}>
            Nema dostupnih analitika.
          </p>
        </div>
      )}
    </div>
  );
};

export default Analytics;
