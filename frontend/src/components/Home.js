import React, { useState, useEffect, useCallback } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import api from '../services/api';

const Home = () => {
  const { isAuthenticated, user } = useAuth();
  const location = useLocation();
  const [recommendations, setRecommendations] = useState(null);
  const [mostPlayedSongs, setMostPlayedSongs] = useState([]);
  const [loading, setLoading] = useState(false);
  const [loadingMostPlayed, setLoadingMostPlayed] = useState(false);
  const [error, setError] = useState(null);

  const fetchRecommendations = useCallback(async () => {
    setLoading(true);
    setError(null);
    
    try {
      // Koristi API Gateway endpoint umesto direktnog poziva na ratings-service
      const token = localStorage.getItem('token');
      const apiUrl = process.env.REACT_APP_API_URL || 'http://localhost:8081';
      const response = await fetch(`${apiUrl}/api/ratings/recommendations?userId=${user.id}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });
      
      if (!response.ok) {
        // Handle specific error cases
        if (response.status === 503) {
          throw new Error('Servis za preporuke trenutno nije dostupan. Molimo pokušajte kasnije.');
        } else if (response.status === 401) {
          throw new Error('Niste autorizovani. Molimo prijavite se ponovo.');
        } else if (response.status === 500) {
          throw new Error('Greška na serveru. Molimo pokušajte kasnije.');
        } else {
          throw new Error(`Greška pri učitavanju preporuka (${response.status})`);
        }
      }
      
      const data = await response.json();
      setRecommendations(data);
    } catch (err) {
      console.error('Error fetching recommendations:', err);
      // Use the error message from the catch block if it's already a user-friendly message
      if (err.message && err.message.includes('Servis') || err.message.includes('Niste') || err.message.includes('Greška')) {
        setError(err.message);
      } else if (err.name === 'TypeError' && err.message.includes('fetch')) {
        setError('Nema konekcije sa serverom. Proverite da li su svi servisi pokrenuti.');
      } else {
        setError('Nije moguće učitati preporuke. Molimo pokušajte kasnije.');
      }
    } finally {
      setLoading(false);
    }
  }, [isAuthenticated, user]);

  useEffect(() => {
    if (isAuthenticated && user?.id && user?.username !== 'admin') {
      fetchRecommendations();
    }
  }, [isAuthenticated, user, fetchRecommendations]);

  // Refresh recommendations when navigating to home page
  useEffect(() => {
    if (location.pathname === '/' && isAuthenticated && user?.id && user?.username !== 'admin') {
      fetchRecommendations();
    }
  }, [location.pathname, isAuthenticated, user, fetchRecommendations]);

  // Refresh recommendations when window gets focus (user might have subscribed elsewhere)
  useEffect(() => {
    const handleFocus = () => {
      if (isAuthenticated && user?.id && user?.username !== 'admin') {
        fetchRecommendations();
      }
    };
    window.addEventListener('focus', handleFocus);
    return () => window.removeEventListener('focus', handleFocus);
  }, [isAuthenticated, user, fetchRecommendations]);

  // Fetch songs from subscribed genres with rating >= 4 (replaces most played songs)
  useEffect(() => {
    const fetchSubscribedGenreSongs = async () => {
      if (!isAuthenticated || !user?.id || user?.username === 'admin') {
        return;
      }
      
      setLoadingMostPlayed(true);
      try {
        const token = localStorage.getItem('token');
        const apiUrl = process.env.REACT_APP_API_URL || 'http://localhost:8081';
        const response = await fetch(`${apiUrl}/api/ratings/recommendations?userId=${user.id}`, {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
          }
        });
        
        if (response.ok) {
          const data = await response.json();
          // Use subscribed genre songs instead of most played
          const songs = data.subscribedGenreSongs || [];
          // Limit to 5 songs (as requested) and add play count info if available
          setMostPlayedSongs(songs.slice(0, 5).map(song => ({
            ...song,
            playCount: song.playCount || 0
          })));
        } else {
          setMostPlayedSongs([]);
        }
      } catch (err) {
        console.error('Error fetching subscribed genre songs:', err);
        setMostPlayedSongs([]);
      } finally {
        setLoadingMostPlayed(false);
      }
    };
    fetchSubscribedGenreSongs();
  }, [isAuthenticated, user]);

  const SongCard = ({ song, reason }) => (
    <div className="song-card-modern" style={{
      background: 'linear-gradient(135deg, rgba(255,255,255,0.95) 0%, rgba(255,255,255,0.9) 100%)',
      backdropFilter: 'blur(10px)',
      borderRadius: '16px',
      padding: '24px',
      marginBottom: '20px',
      border: '1px solid rgba(255,255,255,0.3)',
      boxShadow: '0 8px 32px rgba(0,0,0,0.1)',
      transition: 'all 0.3s ease',
      cursor: 'pointer'
    }}
    onMouseEnter={(e) => {
      e.currentTarget.style.transform = 'translateY(-8px)';
      e.currentTarget.style.boxShadow = '0 12px 40px rgba(102, 126, 234, 0.3)';
    }}
    onMouseLeave={(e) => {
      e.currentTarget.style.transform = 'translateY(0)';
      e.currentTarget.style.boxShadow = '0 8px 32px rgba(0,0,0,0.1)';
    }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginBottom: '16px' }}>
        <div style={{
          width: '50px',
          height: '50px',
          borderRadius: '12px',
          background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          fontSize: '24px',
          color: 'white',
          fontWeight: 'bold'
        }}>
          🎵
        </div>
        <div style={{ flex: 1 }}>
          <h4 style={{ margin: 0, fontSize: '20px', fontWeight: '600', color: '#333' }}>{song.name}</h4>
          <p style={{ margin: '4px 0 0 0', fontSize: '14px', color: '#666' }}>{song.genre}</p>
        </div>
      </div>
      <div style={{ 
        display: 'grid', 
        gridTemplateColumns: 'repeat(2, 1fr)', 
        gap: '12px', 
        marginBottom: '16px',
        padding: '12px',
        background: 'rgba(102, 126, 234, 0.05)',
        borderRadius: '8px'
      }}>
        <div>
          <span style={{ fontSize: '12px', color: '#666', fontWeight: '500' }}>⏱️ Trajanje:</span>
          <span style={{ marginLeft: '8px', fontWeight: '600', color: '#333' }}>
            {Math.floor(song.duration / 60)}:{(song.duration % 60).toString().padStart(2, '0')}
          </span>
        </div>
        <div>
          <span style={{ fontSize: '12px', color: '#666', fontWeight: '500' }}>💡 Razlog:</span>
          <span style={{ marginLeft: '8px', fontWeight: '600', color: '#667eea' }}>{reason}</span>
        </div>
      </div>
      <Link 
        to={`/songs/${song.songId}`} 
        className="btn btn-primary"
        style={{
          width: '100%',
          textAlign: 'center',
          display: 'inline-block',
          textDecoration: 'none',
          padding: '12px 24px',
          borderRadius: '8px',
          fontWeight: '600'
        }}
      >
        🎧 Slušaj pesmu
      </Link>
    </div>
  );

  return (
    <div style={{ minHeight: 'calc(100vh - 80px)' }}>
      {!isAuthenticated ? (
        // Hero Section for non-authenticated users
        <div style={{
          background: 'linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%)',
          padding: '80px 20px',
          textAlign: 'center'
        }}>
          <div className="container" style={{ maxWidth: '800px' }}>
            <div style={{
              fontSize: '72px',
              marginBottom: '20px',
              animation: 'float 3s ease-in-out infinite'
            }}>
              🎵🎶🎧
            </div>
            <h1 style={{
              fontSize: '48px',
              fontWeight: '700',
              marginBottom: '20px',
              background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
              WebkitBackgroundClip: 'text',
              WebkitTextFillColor: 'transparent',
              backgroundClip: 'text'
            }}>
              Dobrodošli u Music Streaming
            </h1>
            <p style={{
              fontSize: '20px',
              color: '#666',
              marginBottom: '40px',
              lineHeight: '1.6'
            }}>
              Otkrijte milion pesama, slušajte svoje omiljene izvođače i kreirajte savršene playliste
            </p>
            <div style={{ display: 'flex', gap: '20px', justifyContent: 'center', flexWrap: 'wrap' }}>
              <Link 
                to="/login" 
                className="btn btn-primary"
                style={{
                  padding: '16px 32px',
                  fontSize: '18px',
                  fontWeight: '600',
                  borderRadius: '12px',
                  textDecoration: 'none',
                  display: 'inline-block'
                }}
              >
                🚀 Prijavi se
              </Link>
              <Link 
                to="/register" 
                className="btn btn-secondary"
                style={{
                  padding: '16px 32px',
                  fontSize: '18px',
                  fontWeight: '600',
                  borderRadius: '12px',
                  textDecoration: 'none',
                  display: 'inline-block',
                  background: 'white',
                  color: '#667eea',
                  border: '2px solid #667eea'
                }}
              >
                ✨ Registruj se
              </Link>
            </div>
          </div>
        </div>
      ) : (
        // Authenticated user content
        <div className="container" style={{ paddingTop: '40px', paddingBottom: '40px' }}>
          {/* Welcome Card */}
          <div style={{
            background: 'linear-gradient(135deg, rgba(255,255,255,0.95) 0%, rgba(255,255,255,0.9) 100%)',
            backdropFilter: 'blur(10px)',
            borderRadius: '20px',
            padding: '40px',
            marginBottom: '40px',
            boxShadow: '0 10px 40px rgba(0,0,0,0.1)',
            border: '1px solid rgba(255,255,255,0.3)'
          }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '20px', marginBottom: '30px' }}>
              <div style={{
                width: '80px',
                height: '80px',
                borderRadius: '20px',
                background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontSize: '40px',
                boxShadow: '0 8px 20px rgba(102, 126, 234, 0.3)'
              }}>
                👋
              </div>
              <div>
                <h1 style={{ margin: 0, fontSize: '32px', fontWeight: '700', color: '#333' }}>
                  Zdravo, {user.username}!
                </h1>
                <p style={{ margin: '8px 0 0 0', fontSize: '16px', color: '#666' }}>
                  {user?.username === 'admin' ? 'Admin panel' : 'Vaša muzička biblioteka'}
                </p>
              </div>
            </div>

            {/* Quick Actions */}
            <div style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
              gap: '16px',
              marginTop: '30px'
            }}>
              <Link 
                to="/artists" 
                className="quick-action-card"
                style={{
                  padding: '20px',
                  background: 'linear-gradient(135deg, #667eea15 0%, #764ba215 100%)',
                  borderRadius: '12px',
                  textDecoration: 'none',
                  color: '#333',
                  border: '2px solid transparent',
                  transition: 'all 0.3s ease',
                  textAlign: 'center'
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.borderColor = '#667eea';
                  e.currentTarget.style.transform = 'translateY(-4px)';
                  e.currentTarget.style.boxShadow = '0 8px 20px rgba(102, 126, 234, 0.2)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.borderColor = 'transparent';
                  e.currentTarget.style.transform = 'translateY(0)';
                  e.currentTarget.style.boxShadow = 'none';
                }}
              >
                <div style={{ fontSize: '32px', marginBottom: '8px' }}>🎤</div>
                <div style={{ fontWeight: '600', fontSize: '16px' }}>Izvođači</div>
              </Link>
              <Link 
                to="/albums" 
                className="quick-action-card"
                style={{
                  padding: '20px',
                  background: 'linear-gradient(135deg, #667eea15 0%, #764ba215 100%)',
                  borderRadius: '12px',
                  textDecoration: 'none',
                  color: '#333',
                  border: '2px solid transparent',
                  transition: 'all 0.3s ease',
                  textAlign: 'center'
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.borderColor = '#667eea';
                  e.currentTarget.style.transform = 'translateY(-4px)';
                  e.currentTarget.style.boxShadow = '0 8px 20px rgba(102, 126, 234, 0.2)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.borderColor = 'transparent';
                  e.currentTarget.style.transform = 'translateY(0)';
                  e.currentTarget.style.boxShadow = 'none';
                }}
              >
                <div style={{ fontSize: '32px', marginBottom: '8px' }}>💿</div>
                <div style={{ fontWeight: '600', fontSize: '16px' }}>Albumi</div>
              </Link>
              <Link 
                to="/songs" 
                className="quick-action-card"
                style={{
                  padding: '20px',
                  background: 'linear-gradient(135deg, #667eea15 0%, #764ba215 100%)',
                  borderRadius: '12px',
                  textDecoration: 'none',
                  color: '#333',
                  border: '2px solid transparent',
                  transition: 'all 0.3s ease',
                  textAlign: 'center'
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.borderColor = '#667eea';
                  e.currentTarget.style.transform = 'translateY(-4px)';
                  e.currentTarget.style.boxShadow = '0 8px 20px rgba(102, 126, 234, 0.2)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.borderColor = 'transparent';
                  e.currentTarget.style.transform = 'translateY(0)';
                  e.currentTarget.style.boxShadow = 'none';
                }}
              >
                <div style={{ fontSize: '32px', marginBottom: '8px' }}>🎵</div>
                <div style={{ fontWeight: '600', fontSize: '16px' }}>Pesme</div>
              </Link>
              {user && (
                <Link 
                  to={`/notifications?userId=${user.id}`} 
                  className="quick-action-card"
                  style={{
                    padding: '20px',
                    background: 'linear-gradient(135deg, #667eea15 0%, #764ba215 100%)',
                    borderRadius: '12px',
                    textDecoration: 'none',
                    color: '#333',
                    border: '2px solid transparent',
                    transition: 'all 0.3s ease',
                    textAlign: 'center'
                  }}
                  onMouseEnter={(e) => {
                    e.currentTarget.style.borderColor = '#667eea';
                    e.currentTarget.style.transform = 'translateY(-4px)';
                    e.currentTarget.style.boxShadow = '0 8px 20px rgba(102, 126, 234, 0.2)';
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.borderColor = 'transparent';
                    e.currentTarget.style.transform = 'translateY(0)';
                    e.currentTarget.style.boxShadow = 'none';
                  }}
                >
                  <div style={{ fontSize: '32px', marginBottom: '8px' }}>🔔</div>
                  <div style={{ fontWeight: '600', fontSize: '16px' }}>Notifikacije</div>
                </Link>
              )}
            </div>
          </div>

          {/* Most Played Songs Section (2.12) - shown to ALL users */}
          {mostPlayedSongs.length > 0 && (
            <div style={{
              background: 'linear-gradient(135deg, rgba(255,255,255,0.95) 0%, rgba(255,255,255,0.9) 100%)',
              backdropFilter: 'blur(10px)',
              borderRadius: '20px',
              padding: '40px',
              marginBottom: '40px',
              boxShadow: '0 10px 40px rgba(0,0,0,0.1)',
              border: '1px solid rgba(255,255,255,0.3)'
            }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '16px', marginBottom: '30px' }}>
                <div style={{ fontSize: '40px' }}>🔥</div>
                <h2 style={{ margin: 0, fontSize: '28px', fontWeight: '700', color: '#333' }}>
                  Pesme iz žanrova na koje ste pretplaćeni
                </h2>
              </div>
              {loadingMostPlayed ? (
                <div style={{ textAlign: 'center', padding: '20px' }}>
                  <p>Učitavanje...</p>
                </div>
              ) : (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                  {mostPlayedSongs.map((song, index) => (
                    <div
                      key={song.songId || song.id || index}
                      style={{
                        padding: '16px',
                        background: 'linear-gradient(135deg, rgba(255,255,255,0.95) 0%, rgba(255,255,255,0.9) 100%)',
                        borderRadius: '12px',
                        border: '1px solid rgba(102, 126, 234, 0.2)',
                        display: 'flex',
                        alignItems: 'center',
                        gap: '16px',
                        cursor: 'pointer',
                        transition: 'all 0.3s ease',
                        boxShadow: '0 2px 8px rgba(0,0,0,0.1)'
                      }}
                      onMouseEnter={(e) => {
                        e.currentTarget.style.transform = 'translateX(4px)';
                        e.currentTarget.style.boxShadow = '0 4px 12px rgba(102, 126, 234, 0.2)';
                      }}
                      onMouseLeave={(e) => {
                        e.currentTarget.style.transform = 'translateX(0)';
                        e.currentTarget.style.boxShadow = '0 2px 8px rgba(0,0,0,0.1)';
                      }}
                      onClick={() => window.location.href = `/songs/${song.songId || song.id}`}
                    >
                      <div style={{
                        width: '50px',
                        height: '50px',
                        borderRadius: '10px',
                        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        color: 'white',
                        fontWeight: '700',
                        fontSize: '18px',
                        flexShrink: 0
                      }}>
                        {index + 1}
                      </div>
                      <div style={{ flex: 1 }}>
                        <div style={{ fontWeight: '600', fontSize: '16px', color: '#333', marginBottom: '4px' }}>
                          {song.name}
                        </div>
                        <div style={{ fontSize: '14px', color: '#666' }}>
                          {song.genre} {song.playCount ? `• ${song.playCount} puštanja` : ''}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* Admin Panel */}
          {user?.username === 'admin' && (
            <div style={{
              background: 'linear-gradient(135deg, rgba(255, 193, 7, 0.1) 0%, rgba(255, 152, 0, 0.1) 100%)',
              borderRadius: '20px',
              padding: '40px',
              marginBottom: '40px',
              border: '2px solid rgba(255, 193, 7, 0.3)'
            }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '16px', marginBottom: '20px' }}>
                <div style={{ fontSize: '40px' }}>⚙️</div>
                <h2 style={{ margin: 0, fontSize: '28px', fontWeight: '700', color: '#333' }}>Admin Panel</h2>
              </div>
              <p style={{ fontSize: '16px', color: '#666', lineHeight: '1.6' }}>
                Dobrodošli u admin dashboard. Imate pristup svim administrativnim funkcijama za upravljanje izvođačima, albumima i pesmama.
              </p>
            </div>
          )}

          {/* Recommendations Section */}
          {user?.username !== 'admin' && (
            <div style={{
              background: 'linear-gradient(135deg, rgba(255,255,255,0.95) 0%, rgba(255,255,255,0.9) 100%)',
              backdropFilter: 'blur(10px)',
              borderRadius: '20px',
              padding: '40px',
              marginBottom: '40px',
              boxShadow: '0 10px 40px rgba(0,0,0,0.1)',
              border: '1px solid rgba(255,255,255,0.3)'
            }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '16px', marginBottom: '30px' }}>
                <div style={{ fontSize: '40px' }}>✨</div>
                <h2 style={{ margin: 0, fontSize: '28px', fontWeight: '700', color: '#333' }}>
                  Personalizovane Preporuke
                </h2>
              </div>

              {loading && (
                <div style={{ textAlign: 'center', padding: '40px' }}>
                  <div style={{ fontSize: '48px', marginBottom: '16px', animation: 'spin 1s linear infinite' }}>⏳</div>
                  <p style={{ color: '#666', fontSize: '16px' }}>Učitavanje preporuka...</p>
                </div>
              )}

              {error && (
                <div style={{
                  background: 'linear-gradient(135deg, rgba(255, 193, 7, 0.1) 0%, rgba(255, 152, 0, 0.1) 100%)',
                  border: '2px solid rgba(255, 193, 7, 0.3)',
                  borderRadius: '12px',
                  padding: '24px',
                  marginBottom: '20px',
                  textAlign: 'center'
                }}>
                  <div style={{ fontSize: '48px', marginBottom: '16px' }}>⚠️</div>
                  <p style={{ 
                    color: '#666', 
                    fontSize: '16px', 
                    lineHeight: '1.6',
                    marginBottom: '20px'
                  }}>
                    {error}
                  </p>
                  <button
                    onClick={fetchRecommendations}
                    style={{
                      padding: '12px 24px',
                      fontSize: '16px',
                      fontWeight: '600',
                      borderRadius: '8px',
                      background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                      color: 'white',
                      border: 'none',
                      cursor: 'pointer',
                      transition: 'all 0.3s ease'
                    }}
                    onMouseEnter={(e) => {
                      e.currentTarget.style.transform = 'translateY(-2px)';
                      e.currentTarget.style.boxShadow = '0 4px 12px rgba(102, 126, 234, 0.3)';
                    }}
                    onMouseLeave={(e) => {
                      e.currentTarget.style.transform = 'translateY(0)';
                      e.currentTarget.style.boxShadow = 'none';
                    }}
                  >
                    🔄 Pokušaj ponovo
                  </button>
                </div>
              )}


              {recommendations && (
                <>
                  {recommendations.topRatedSong && (
                    <div style={{ marginBottom: '40px' }}>
                      <h3 style={{
                        fontSize: '22px',
                        fontWeight: '600',
                        marginBottom: '20px',
                        color: '#333',
                        display: 'flex',
                        alignItems: 'center',
                        gap: '10px'
                      }}>
                        <span>🔍</span>
                        <span>Otkrijte nešto novo</span>
                      </h3>
                      <SongCard song={recommendations.topRatedSong} reason={recommendations.topRatedSong.reason} />
                    </div>
                  )}

                  {!recommendations.topRatedSong && (
                    <div style={{
                      textAlign: 'center',
                      padding: '40px',
                      background: 'rgba(102, 126, 234, 0.05)',
                      borderRadius: '12px'
                    }}>
                      <div style={{ fontSize: '48px', marginBottom: '16px' }}>🎵</div>
                      <p style={{ color: '#666', fontSize: '16px', lineHeight: '1.6' }}>
                        Nema dostupnih preporuka. Počnite da se pretplaćujete na žanrove i ocenjujete pesme da biste dobili personalizovane preporuke!
                      </p>
                    </div>
                  )}
                </>
              )}
            </div>
          )}
        </div>
      )}
      
      <style>{`
        @keyframes float {
          0%, 100% { transform: translateY(0px); }
          50% { transform: translateY(-20px); }
        }
        @keyframes spin {
          from { transform: rotate(0deg); }
          to { transform: rotate(360deg); }
        }
      `}</style>
    </div>
  );
};

export default Home;
