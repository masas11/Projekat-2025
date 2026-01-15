import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import api from '../services/api';

const SongDetail = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [song, setSong] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

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
      </div>
    </div>
  );
};

export default SongDetail;
