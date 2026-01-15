import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import api from '../services/api';

const Albums = () => {
  const [albums, setAlbums] = useState([]);
  const [artists, setArtists] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showForm, setShowForm] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    releaseDate: '',
    genre: '',
    artistIDs: '',
  });
  const { isAdmin } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    loadAlbums();
    loadArtists();
  }, []);

  const loadAlbums = async () => {
    try {
      const data = await api.getAlbums();
      setAlbums(Array.isArray(data) ? data : []);
    } catch (err) {
      setError(err.message || 'Greška pri učitavanju albuma');
    } finally {
      setLoading(false);
    }
  };

  const loadArtists = async () => {
    try {
      const data = await api.getArtists();
      setArtists(Array.isArray(data) ? data : []);
    } catch (err) {
      console.error('Error loading artists:', err);
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
    setError();

    const albumData = {
      name: formData.name,
      releaseDate: formData.releaseDate,
      genre: formData.genre,
      artistIDs: formData.artistIDs.split(',').map(id => id.trim()).filter(id => id),
    };

    try {
      await api.createAlbum(albumData);
      setShowForm(false);
      setFormData({ name: '', releaseDate: '', genre: '', artistIDs: '' });
      loadAlbums();
    } catch (err) {
      setError(err.message || 'Greška pri čuvanju albuma');
    }
  };

  if (loading) {
    return <div className="container">Učitavanje...</div>;
  }

  return (
    <div className="container">
      <div className="card">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <h2>Albumi</h2>
          {isAdmin() && (
            <button className="btn btn-primary" onClick={() => setShowForm(!showForm)}>
              {showForm ? 'Otkaži' : 'Dodaj album'}
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
              <label>Datum izdavanja:</label>
              <input
                type="date"
                name="releaseDate"
                value={formData.releaseDate}
                onChange={handleChange}
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
          {albums.length === 0 ? (
            <p>Nema albuma.</p>
          ) : (
            albums.map((album) => (
              <div
                key={album.id}
                className="list-item"
                onClick={() => navigate(`/albums/${album.id}`)}
              >
                <h3>{album.name}</h3>
                {album.genre && <span className="genre-tag">{album.genre}</span>}
                {album.releaseDate && (
                  <p style={{ marginTop: '5px' }}>
                    Datum izdavanja: {new Date(album.releaseDate).toLocaleDateString()}
                  </p>
                )}
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
};

export default Albums;
