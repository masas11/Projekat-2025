import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import api from '../services/api';

const Songs = () => {
  const [songs, setSongs] = useState([]);
  const [albums, setAlbums] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showForm, setShowForm] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    duration: '',
    genre: '',
    albumID: '',
    artistIDs: '',
  });
  const { isAdmin } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    loadSongs();
    loadAlbums();
  }, []);

  const loadSongs = async () => {
    try {
      const data = await api.getSongs();
      setSongs(Array.isArray(data) ? data : []);
    } catch (err) {
      setError(err.message || 'Greška pri učitavanju pesama');
    } finally {
      setLoading(false);
    }
  };

  const loadAlbums = async () => {
    try {
      const data = await api.getAlbums();
      setAlbums(Array.isArray(data) ? data : []);
    } catch (err) {
      console.error('Error loading albums:', err);
    }
  };

  const handleChange = (e) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');

    const songData = {
      name: formData.name,
      duration: parseInt(formData.duration),
      genre: formData.genre,
      albumID: formData.albumID,
      artistIDs: formData.artistIDs.split(',').map(id => id.trim()).filter(id => id),
    };

    try {
      await api.createSong(songData);
      setShowForm(false);
      setFormData({ name: '', duration: '', genre: '', albumID: '', artistIDs: '' });
      loadSongs();
    } catch (err) {
      setError(err.message || 'Greška pri čuvanju pesme');
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

  return (
    <div className="container">
      <div className="card">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <h2>Pesme</h2>
          {isAdmin() && (
            <button className="btn btn-primary" onClick={() => setShowForm(!showForm)}>
              {showForm ? 'Otkaži' : 'Dodaj pesmu'}
            </button>
          )}
        </div>

        {showForm && isAdmin() && (
          <form onSubmit={handleSubmit} style={{ marginTop: '20px' }}>
            <div className="form-group">
              <label>Naziv:</label>
              <input
                type="text"
                name="name"
                value={formData.name}
                onChange={handleChange}
                required
              />
            </div>
            <div className="form-group">
              <label>Trajanje (u sekundama):</label>
              <input
                type="number"
                name="duration"
                value={formData.duration}
                onChange={handleChange}
                required
                min="1"
              />
            </div>
            <div className="form-group">
              <label>Žanr:</label>
              <input
                type="text"
                name="genre"
                value={formData.genre}
                onChange={handleChange}
                required
              />
            </div>
            <div className="form-group">
              <label>ID albuma:</label>
              <input
                type="text"
                name="albumID"
                value={formData.albumID}
                onChange={handleChange}
                required
              />
            </div>
            <div className="form-group">
              <label>ID izvođača (odvojeni zarezom):</label>
              <input
                type="text"
                name="artistIDs"
                value={formData.artistIDs}
                onChange={handleChange}
                placeholder="id1, id2, id3"
                required
              />
            </div>
            {error && <div className="error">{error}</div>}
            <div style={{ display: 'flex', gap: '10px' }}>
              <button type="submit" className="btn btn-primary">Dodaj</button>
              <button type="button" className="btn btn-secondary" onClick={() => setShowForm(false)}>
                Otkaži
              </button>
            </div>
          </form>
        )}

        {error && !showForm && <div className="error">{error}</div>}

        <div style={{ marginTop: '20px' }}>
          {songs.length === 0 ? (
            <p>Nema pesama.</p>
          ) : (
            songs.map((song) => (
              <div
                key={song.id}
                className="list-item"
                onClick={() => navigate(`/songs/${song.id}`)}
              >
                <h3>{song.name}</h3>
                {song.duration && <p>Trajanje: {formatDuration(song.duration)}</p>}
                {song.genre && <span className="genre-tag">{song.genre}</span>}
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
};

export default Songs;
