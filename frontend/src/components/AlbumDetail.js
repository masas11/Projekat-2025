import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import api from '../services/api';

const AlbumDetail = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [album, setAlbum] = useState(null);
  const [songs, setSongs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    loadAlbum();
    loadSongs();
  }, [id]);

  const loadAlbum = async () => {
    try {
      const data = await api.getAlbum(id);
      setAlbum(data);
    } catch (err) {
      setError(err.message || 'Greška pri učitavanju albuma');
    } finally {
      setLoading(false);
    }
  };

  const loadSongs = async () => {
    try {
      const data = await api.getSongsByAlbum(id);
      setSongs(Array.isArray(data) ? data : []);
    } catch (err) {
      console.error('Error loading songs:', err);
    }
  };

  if (loading) {
    return <div className="container">Učitavanje...</div>;
  }

  if (error || !album) {
    return (
      <div className="container">
        <div className="card">
          <div className="error">{error || 'Album nije pronađen'}</div>
          <button className="btn btn-secondary" onClick={() => navigate('/albums')}>
            Nazad
          </button>
        </div>
      </div>
    );
  }

  const formatDuration = (seconds) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  return (
    <div className="container">
      <div className="card">
        <button className="btn btn-secondary" onClick={() => navigate('/albums')} style={{ marginBottom: '20px' }}>
          ← Nazad na albume
        </button>
        <h2>{album.name}</h2>
        {album.genre && (
          <div style={{ marginTop: '10px', marginBottom: '10px' }}>
            <span className="genre-tag">{album.genre}</span>
          </div>
        )}
        {album.releaseDate && (
          <p style={{ marginTop: '10px' }}>
            Datum izdavanja: {new Date(album.releaseDate).toLocaleDateString()}
          </p>
        )}
      </div>

      {songs.length > 0 && (
        <div className="card">
          <h3>Pesme</h3>
          {songs.map((song) => (
            <div
              key={song.id}
              className="list-item"
              onClick={() => navigate(`/songs/${song.id}`)}
            >
              <h4>{song.name}</h4>
              {song.duration && <p>Trajanje: {formatDuration(song.duration)}</p>}
              {song.genre && <span className="genre-tag">{song.genre}</span>}
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default AlbumDetail;
