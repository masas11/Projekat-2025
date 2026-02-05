import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import api from '../services/api';
import AudioPlayer from './AudioPlayer';

const SongDetail = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const { user } = useAuth();
  const [song, setSong] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [rating, setRating] = useState(0);
  const [hoveredStar, setHoveredStar] = useState(0);
  const [ratingMessage, setRatingMessage] = useState('');

  useEffect(() => {
    loadSong();
  }, [id]);

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

  const formatDuration = (seconds) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  const rateSong = async (ratingValue) => {
    if (!user) {
      setRatingMessage('Morate biti prijavljeni da biste ocenili pesmu');
      return;
    }

    try {
      // Call the ratings service directly
      const response = await fetch(`http://localhost:8003/rate-song?songId=${id}&rating=${ratingValue}&userId=${user.id}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        setRating(ratingValue);
        setRatingMessage('Uspešno ste ocenili pesmu!');
      } else {
        const errorText = await response.text();
        setRatingMessage(`Greška: ${errorText}`);
      }
    } catch (err) {
      setRatingMessage(`Greška pri ocenjivanju: ${err.message}`);
    }
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
        <button className="btn btn-secondary" onClick={() => navigate('/songs')} style={{ marginBottom: '20px' }}>
          ← Nazad na pesme
        </button>
        <h2>{song.name}</h2>
        {song.duration && (
          <p style={{ marginTop: '10px', marginBottom: '10px' }}>
            Trajanje: {formatDuration(song.duration)}
          </p>
        )}
        {song.genre && (
          <div style={{ marginTop: '10px', marginBottom: '10px' }}>
            <span className="genre-tag">{song.genre}</span>
          </div>
        )}
        {song.albumID && (
          <p style={{ marginTop: '10px' }}>
            <button
              className="btn btn-secondary"
              onClick={() => navigate(`/albums/${song.albumID}`)}
            >
              Vidi album
            </button>
          </p>
        )}
        
        {/* Audio Player */}
        <div style={{ marginTop: '30px' }}>
          <AudioPlayer 
            songId={song.id} 
            songName={song.name} 
            audioFileUrl={song.audioFileUrl} 
          />
        </div>

        {/* Rating Section */}
        <div style={{ marginTop: '30px', padding: '20px', backgroundColor: '#f8f9fa', borderRadius: '5px' }}>
          <h4>Oceni pesmu</h4>
          {user ? (
            <div>
              <div style={{ fontSize: '24px', marginBottom: '10px' }}>
                {[1, 2, 3, 4, 5].map((star) => (
                  <span
                    key={star}
                    style={{
                      cursor: 'pointer',
                      color: star <= (hoveredStar || rating) ? '#ffc107' : '#ddd',
                      marginRight: '5px'
                    }}
                    onClick={() => rateSong(star)}
                    onMouseEnter={() => setHoveredStar(star)}
                    onMouseLeave={() => setHoveredStar(0)}
                  >
                    ★
                  </span>
                ))}
              </div>
              {ratingMessage && (
                <div style={{ 
                  marginTop: '10px', 
                  padding: '10px', 
                  backgroundColor: ratingMessage.includes('Uspešno') ? '#d4edda' : '#f8d7da',
                  borderRadius: '4px',
                  fontSize: '0.9em'
                }}>
                  {ratingMessage}
                </div>
              )}
            </div>
          ) : (
            <p style={{ color: '#666' }}>
              <a href="/login" style={{ color: '#007bff' }}>Prijavite se</a> da biste ocenili pesmu
            </p>
          )}
        </div>
      </div>
    </div>
  );
};

export default SongDetail;
