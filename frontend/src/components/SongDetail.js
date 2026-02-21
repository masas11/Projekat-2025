import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import api from '../services/api';
import AudioPlayer from './AudioPlayer';

const SongDetail = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const { user, isAuthenticated, isAdmin } = useAuth();
  const [song, setSong] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [userRating, setUserRating] = useState(null);
  const [ratingMessage, setRatingMessage] = useState('');
  const [isRating, setIsRating] = useState(false);

  useEffect(() => {
    loadSong();
    if (isAuthenticated && user) {
      loadUserRating();
    }
  }, [id, isAuthenticated, user]);

  const loadSong = async () => {
    try {
      const data = await api.getSong(id);
      setSong(data);
    } catch (err) {
      setError(err.message || 'Greška pri učitavanju pesme');
    } finally {
      setLoading(false);
    }
  };

  const loadUserRating = async () => {
    if (!isAuthenticated || !user) return;

    try {
      const response = await api.getRating(id, user.id);
      if (response && response.rating !== null && response.rating !== undefined) {
        setUserRating(response.rating);
      } else {
        setUserRating(null);
      }
    } catch (err) {
      // Rating doesn't exist, which is fine
      setUserRating(null);
    }
  };

  const handleRateSong = async (rating) => {
    if (!isAuthenticated || !user) {
      setError('Morate biti prijavljeni da biste ocenili pesmu');
      return;
    }

    if (isAdmin()) {
      setError('Administratori ne mogu da ocenjuju pesme');
      return;
    }

    setIsRating(true);
    setRatingMessage('');
    setError('');

    try {
      await api.rateSong(id, rating, user.id);
      setUserRating(rating);
      setRatingMessage(`Uspešno ste ocenili pesmu sa ocenom: ${rating}!`);
      setTimeout(() => setRatingMessage(''), 3000);
    } catch (err) {
      setError(err.message || 'Greška pri ocenjivanju pesme');
    } finally {
      setIsRating(false);
    }
  };

  const handleDeleteRating = async () => {
    if (!isAuthenticated || !user) {
      setError('Morate biti prijavljeni da biste obrisali ocenu');
      return;
    }

    if (!window.confirm('Da li ste sigurni da želite da obrišete ovu ocenu?')) {
      return;
    }

    setIsRating(true);
    setRatingMessage('');
    setError('');

    try {
      await api.deleteRating(id, user.id);
      setUserRating(null);
      setRatingMessage('Ocena je uspešno obrisana!');
      setTimeout(() => setRatingMessage(''), 3000);
    } catch (err) {
      setError(err.message || 'Greška pri brisanju ocene');
    } finally {
      setIsRating(false);
    }
  };

  const renderRatingStars = () => {
    if (!isAuthenticated) {
      return (
        <div style={{ 
          padding: '30px', 
          background: 'linear-gradient(135deg, rgba(102, 126, 234, 0.05) 0%, rgba(118, 75, 162, 0.05) 100%)',
          borderRadius: '16px', 
          textAlign: 'center',
          border: '2px solid rgba(102, 126, 234, 0.2)'
        }}>
          <div style={{ fontSize: '48px', marginBottom: '16px' }}>🔒</div>
          <p style={{ margin: '0', fontSize: '16px', color: '#666', fontWeight: '500' }}>
            Morate biti prijavljeni da biste ocenili pesmu
          </p>
        </div>
      );
    }

    if (isAdmin()) {
      return null; // Ne prikazuj ništa za admin korisnike
    }

    return (
      <>
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginBottom: '24px' }}>
          <span style={{ fontSize: '32px' }}>⭐</span>
          <h2 style={{
            margin: 0,
            fontSize: '24px',
            fontWeight: '700',
            color: '#333'
          }}>
            Oceni pesmu
          </h2>
          {userRating && (
            <span style={{ 
              marginLeft: 'auto',
              padding: '8px 16px',
              background: 'linear-gradient(135deg, rgba(40, 167, 69, 0.1) 0%, rgba(40, 167, 69, 0.15) 100%)',
              borderRadius: '12px',
              fontSize: '16px', 
              color: '#28a745', 
              fontWeight: '600',
              border: '1px solid rgba(40, 167, 69, 0.3)'
            }}>
              Vaša ocena: {userRating}/5
            </span>
          )}
        </div>
        
        <div style={{ display: 'flex', gap: '12px', alignItems: 'center', justifyContent: 'center', flexWrap: 'wrap', marginBottom: '20px' }}>
          {[1, 2, 3, 4, 5].map((star) => (
            <button
              key={star}
              onClick={() => handleRateSong(star)}
              disabled={isRating}
              style={{ 
                padding: '16px 20px', 
                fontSize: '28px',
                minWidth: '60px',
                height: '60px',
                borderRadius: '12px',
                cursor: isRating ? 'not-allowed' : 'pointer',
                transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                border: userRating >= star ? '2px solid #667eea' : '2px solid #e0e0e0',
                background: userRating >= star 
                  ? 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)' 
                  : 'white',
                color: userRating >= star ? '#fff' : '#666',
                boxShadow: userRating >= star 
                  ? '0 4px 12px rgba(102, 126, 234, 0.3)' 
                  : '0 2px 4px rgba(0,0,0,0.1)'
              }}
              title={`Oceni sa ${star} zvezdic${star === 1 ? 'u' : star >= 4 ? 'e' : 'a'}`}
              onMouseEnter={(e) => {
                if (!isRating) {
                  e.target.style.transform = 'scale(1.15) translateY(-4px)';
                  e.target.style.boxShadow = '0 8px 20px rgba(102, 126, 234, 0.4)';
                  if (userRating < star) {
                    e.target.style.background = 'linear-gradient(135deg, rgba(102, 126, 234, 0.2) 0%, rgba(118, 75, 162, 0.2) 100%)';
                    e.target.style.borderColor = '#667eea';
                  }
                }
              }}
              onMouseLeave={(e) => {
                if (!isRating) {
                  e.target.style.transform = 'scale(1) translateY(0)';
                  if (userRating < star) {
                    e.target.style.background = 'white';
                    e.target.style.borderColor = '#e0e0e0';
                    e.target.style.boxShadow = '0 2px 4px rgba(0,0,0,0.1)';
                  } else {
                    e.target.style.boxShadow = '0 4px 12px rgba(102, 126, 234, 0.3)';
                  }
                }
              }}
            >
              {userRating >= star ? '★' : '☆'}
            </button>
          ))}
          
          {userRating && (
            <button
              className="btn btn-danger"
              onClick={handleDeleteRating}
              disabled={isRating}
              style={{ 
                padding: '16px 24px', 
                fontSize: '16px',
                height: '60px',
                borderRadius: '12px',
                cursor: isRating ? 'not-allowed' : 'pointer',
                transition: 'all 0.3s ease',
                fontWeight: '600',
                display: 'flex',
                alignItems: 'center',
                gap: '8px'
              }}
              title="Obriši svoju ocenu"
              onMouseEnter={(e) => {
                if (!isRating) {
                  e.target.style.transform = 'scale(1.05)';
                }
              }}
              onMouseLeave={(e) => {
                if (!isRating) {
                  e.target.style.transform = 'scale(1)';
                }
              }}
            >
              {isRating ? '⏳ Brisanje...' : '🗑️ Obriši ocenu'}
            </button>
          )}
        </div>
        
        {isRating && (
          <div style={{ 
            marginTop: '20px', 
            padding: '16px',
            background: 'rgba(102, 126, 234, 0.05)',
            borderRadius: '12px',
            fontSize: '16px', 
            color: '#667eea',
            fontWeight: '500',
            textAlign: 'center',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: '8px'
          }}>
            <span>⏳</span>
            <span>Čuvanje ocene...</span>
          </div>
        )}
      </>
    );
  };

  const formatDuration = (seconds) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  if (loading) {
    return (
      <div className="container" style={{ paddingTop: '60px', textAlign: 'center' }}>
        <div style={{ fontSize: '48px', marginBottom: '20px', animation: 'spin 1s linear infinite' }}>⏳</div>
        <p style={{ fontSize: '18px', color: '#666' }}>Učitavanje pesme...</p>
      </div>
    );
  }

  if (error || !song) {
    return (
      <div style={{ minHeight: 'calc(100vh - 80px)', paddingTop: '40px', paddingBottom: '40px' }}>
        <div className="container">
          <div style={{
            background: 'linear-gradient(135deg, rgba(255,255,255,0.95) 0%, rgba(255,255,255,0.9) 100%)',
            backdropFilter: 'blur(10px)',
            borderRadius: '20px',
            padding: '40px',
            boxShadow: '0 20px 60px rgba(0,0,0,0.15)',
            border: '1px solid rgba(255,255,255,0.3)',
            textAlign: 'center'
          }}>
            <div className="error" style={{ marginBottom: '20px' }}>
              <span>⚠️</span>
              <span>{error || 'Pesma nije pronađena'}</span>
            </div>
            <button className="btn btn-secondary" onClick={() => navigate('/songs')}>
              ← Nazad na pesme
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div style={{ minHeight: 'calc(100vh - 80px)', paddingTop: '40px', paddingBottom: '40px' }}>
      <div className="container">
        {/* Song Header */}
        <div style={{
          background: 'linear-gradient(135deg, rgba(255,255,255,0.95) 0%, rgba(255,255,255,0.9) 100%)',
          backdropFilter: 'blur(10px)',
          borderRadius: '20px',
          padding: '40px',
          marginBottom: '30px',
          boxShadow: '0 20px 60px rgba(0,0,0,0.15)',
          border: '1px solid rgba(255,255,255,0.3)'
        }}>
          <button 
            className="btn btn-secondary" 
            onClick={() => navigate('/songs')} 
            style={{ 
              marginBottom: '30px',
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
              padding: '12px 20px'
            }}
          >
            ← Nazad na pesme
          </button>

          {error && (
            <div className="error" style={{ marginBottom: '20px' }}>
              <span>⚠️</span>
              <span>{error}</span>
            </div>
          )}
          {ratingMessage && (
            <div className="success" style={{ marginBottom: '20px' }}>
              <span>✅</span>
              <span>{ratingMessage}</span>
            </div>
          )}

          <div style={{ display: 'flex', alignItems: 'flex-start', gap: '30px', flexWrap: 'wrap' }}>
            {/* Song Cover */}
            <div style={{
              width: '200px',
              height: '200px',
              borderRadius: '20px',
              background: `linear-gradient(135deg, 
                hsl(${(song.id.charCodeAt(0) * 137.508) % 360}, 70%, 60%) 0%, 
                hsl(${(song.id.charCodeAt(0) * 137.508 + 60) % 360}, 70%, 50%) 100%)`,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontSize: '80px',
              boxShadow: '0 10px 30px rgba(0,0,0,0.2)',
              flexShrink: 0
            }}>
              🎵
            </div>

            {/* Song Info */}
            <div style={{ flex: 1, minWidth: '300px' }}>
              <h1 style={{
                margin: '0 0 20px 0',
                fontSize: '42px',
                fontWeight: '700',
                background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
                backgroundClip: 'text'
              }}>
                {song.name}
              </h1>

              <div style={{ display: 'flex', flexDirection: 'column', gap: '12px', marginBottom: '20px' }}>
                {song.duration && (
                  <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                    <span style={{ fontSize: '18px' }}>⏱️</span>
                    <span style={{ fontSize: '16px', color: '#666', fontWeight: '500' }}>Trajanje:</span>
                    <span style={{ fontSize: '16px', fontWeight: '600', color: '#333' }}>
                      {formatDuration(song.duration)}
                    </span>
                  </div>
                )}

                {song.genre && (
                  <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                    <span style={{ fontSize: '18px' }}>🎵</span>
                    <span 
                      className="genre-tag"
                      style={{
                        padding: '8px 16px',
                        background: 'linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%)',
                        color: '#667eea',
                        border: '1px solid rgba(102, 126, 234, 0.2)',
                        fontWeight: '600',
                        fontSize: '14px',
                        borderRadius: '16px'
                      }}
                    >
                      {song.genre}
                    </span>
                  </div>
                )}

                {song.albumID && (
                  <div style={{ marginTop: '8px' }}>
                    <button
                      className="btn btn-primary"
                      onClick={() => navigate(`/albums/${song.albumID}`)}
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: '8px',
                        padding: '12px 24px',
                        fontSize: '16px',
                        fontWeight: '600'
                      }}
                    >
                      📀 Vidi album
                    </button>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>

        {/* Audio Player */}
        <div style={{
          background: 'linear-gradient(135deg, rgba(255,255,255,0.95) 0%, rgba(255,255,255,0.9) 100%)',
          backdropFilter: 'blur(10px)',
          borderRadius: '20px',
          padding: '40px',
          marginBottom: '30px',
          boxShadow: '0 20px 60px rgba(0,0,0,0.15)',
          border: '1px solid rgba(255,255,255,0.3)'
        }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginBottom: '24px' }}>
            <span style={{ fontSize: '32px' }}>🎧</span>
            <h2 style={{
              margin: 0,
              fontSize: '24px',
              fontWeight: '700',
              color: '#333'
            }}>
              Slušaj pesmu
            </h2>
          </div>
          <AudioPlayer 
            songId={song.id} 
            songName={song.name} 
            audioFileUrl={song.audioFileUrl} 
          />
        </div>
        
        {/* Rating Section */}
        {renderRatingStars() && (
          <div style={{
            background: 'linear-gradient(135deg, rgba(255,255,255,0.95) 0%, rgba(255,255,255,0.9) 100%)',
            backdropFilter: 'blur(10px)',
            borderRadius: '20px',
            padding: '40px',
            boxShadow: '0 20px 60px rgba(0,0,0,0.15)',
            border: '1px solid rgba(255,255,255,0.3)'
          }}>
            {renderRatingStars()}
          </div>
        )}
      </div>
    </div>
  );
};

export default SongDetail;
