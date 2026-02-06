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
      const response = await api.getRating(id);
      if (response && response.rating !== null) {
        setUserRating(response.rating);
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
        <div style={{ padding: '15px', backgroundColor: '#f8f9fa', borderRadius: '8px', textAlign: 'center' }}>
          <p style={{ margin: '0', fontSize: '1em', color: '#666' }}>
            🔒 Morate biti prijavljeni da biste ocenili pesmu
          </p>
        </div>
      );
    }

    if (isAdmin()) {
      return null; // Ne prikazuj ništa za admin korisnike
    }

    return (
      <div style={{ 
        padding: '20px', 
        backgroundColor: '#f8f9fa', 
        borderRadius: '8px',
        border: '1px solid #e9ecef'
      }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '15px' }}>
          <h3 style={{ margin: '0', fontSize: '1.2em', color: '#495057' }}>
            ⭐ Oceni pesmu
          </h3>
          {userRating && (
            <span style={{ 
              fontSize: '1em', 
              color: '#28a745', 
              fontWeight: 'bold'
            }}>
              Vaša ocena: {userRating}/5
            </span>
          )}
        </div>
        
        <div style={{ display: 'flex', gap: '10px', alignItems: 'center', justifyContent: 'center', flexWrap: 'wrap' }}>
          {[1, 2, 3, 4, 5].map((star) => (
            <button
              key={star}
              className={`btn ${userRating === star ? 'btn-primary' : 'btn-outline-secondary'}`}
              onClick={() => handleRateSong(star)}
              disabled={isRating}
              style={{ 
                padding: '12px 16px', 
                fontSize: '1.5em',
                minWidth: '50px',
                height: '50px',
                borderRadius: '8px',
                cursor: isRating ? 'not-allowed' : 'pointer',
                transition: 'all 0.2s ease',
                border: userRating >= star ? '2px solid #007bff' : '1px solid #6c757d',
                backgroundColor: userRating >= star ? '#007bff' : '#fff',
                color: userRating >= star ? '#fff' : '#6c757d'
              }}
              title={`Oceni sa ${star} zvezdic${star === 1 ? 'u' : star >= 4 ? 'e' : 'a'}`}
              onMouseEnter={(e) => {
                if (!isRating) {
                  e.target.style.transform = 'scale(1.1)';
                  e.target.style.backgroundColor = '#007bff';
                  e.target.style.color = '#fff';
                }
              }}
              onMouseLeave={(e) => {
                if (!isRating && userRating < star) {
                  e.target.style.transform = 'scale(1)';
                  e.target.style.backgroundColor = '#fff';
                  e.target.style.color = '#6c757d';
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
                padding: '12px 16px', 
                fontSize: '1em',
                height: '50px',
                borderRadius: '8px',
                cursor: isRating ? 'not-allowed' : 'pointer',
                transition: 'all 0.2s ease'
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
              {isRating ? 'Brisanje...' : '🗑️ Obriši'}
            </button>
          )}
        </div>
        
        {isRating && (
          <div style={{ 
            marginTop: '15px', 
            fontSize: '1em', 
            color: '#007bff',
            fontStyle: 'italic',
            textAlign: 'center'
          }}>
            ⏳ Čuvanje ocene...
          </div>
        )}
      </div>
    );
  };

  const formatDuration = (seconds) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  if (loading) {
    return <div className="container">Učitavanje...</div>;
  }

  if (error || !song) {
    return (
      <div className="container">
        <div className="card">
          <div className="error">{error || 'Pesma nije pronađena'}</div>
          <button className="btn btn-secondary" onClick={() => navigate('/songs')}>
            Nazad
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="container">
      <div className="card">
        <button 
          className="btn btn-secondary" 
          onClick={() => navigate('/songs')} 
          style={{ marginBottom: '20px' }}
        >
          ← Nazad na pesme
        </button>
        
        {error && <div className="error" style={{ marginBottom: '20px' }}>{error}</div>}
        {ratingMessage && <div className="success" style={{ marginBottom: '20px' }}>{ratingMessage}</div>}
        
        <div style={{ textAlign: 'center', marginBottom: '30px' }}>
          <h1 style={{ margin: '0 0 10px 0', color: '#333' }}>{song.name}</h1>
          
          {song.duration && (
            <p style={{ fontSize: '1.1em', color: '#666', marginBottom: '10px' }}>
              ⏱️ Trajanje: {formatDuration(song.duration)}
            </p>
          )}
          
          {song.genre && (
            <div style={{ marginBottom: '10px' }}>
              <span className="genre-tag" style={{ fontSize: '1em', padding: '6px 12px' }}>
                🎵 {song.genre}
              </span>
            </div>
          )}
          
          {song.albumID && (
            <p style={{ marginTop: '15px' }}>
              <button
                className="btn btn-primary"
                onClick={() => navigate(`/albums/${song.albumID}`)}
                style={{ marginRight: '10px' }}
              >
                📀 Vidi album
              </button>
            </p>
          )}
        </div>
        
        {/* Audio Player */}
        <div style={{ 
          marginTop: '30px', 
          padding: '20px', 
          backgroundColor: '#f8f9fa', 
          borderRadius: '8px',
          border: '1px solid #e9ecef'
        }}>
          <h3 style={{ marginTop: '0', marginBottom: '15px', color: '#495057' }}>
            🎧 Slušaj pesmu
          </h3>
          <AudioPlayer 
            songId={song.id} 
            songName={song.name} 
            audioFileUrl={song.audioFileUrl} 
          />
        </div>
        
        {/* Rating Section */}
        <div style={{ marginTop: '30px' }}>
          {renderRatingStars()}
        </div>
      </div>
    </div>
  );
};

export default SongDetail;
