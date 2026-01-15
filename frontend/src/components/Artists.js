import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import api from '../services/api';

const Artists = () => {
  const [artists, setArtists] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showForm, setShowForm] = useState(false);
  const [editingArtist, setEditingArtist] = useState(null);
  const [formData, setFormData] = useState({
    name: '',
    biography: '',
    genres: '',
  });
  const { isAdmin } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    loadArtists();
  }, []);

  const loadArtists = async () => {
    try {
      const data = await api.getArtists();
      setArtists(Array.isArray(data) ? data : []);
    } catch (err) {
      setError(err.message || 'Greška pri učitavanju izvođača');
    } finally {
      setLoading(false);
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

    const artistData = {
      name: formData.name,
      biography: formData.biography,
      genres: formData.genres.split(',').map(g => g.trim()).filter(g => g),
    };

    try {
      if (editingArtist) {
        await api.updateArtist(editingArtist.id, artistData);
      } else {
        await api.createArtist(artistData);
      }
      setShowForm(false);
      setEditingArtist(null);
      setFormData({ name: '', biography: '', genres: '' });
      loadArtists();
    } catch (err) {
      setError(err.message || 'Greška pri čuvanju izvođača');
    }
  };

  const handleEdit = (artist) => {
    setEditingArtist(artist);
    setFormData({
      name: artist.name,
      biography: artist.biography || '',
      genres: artist.genres ? artist.genres.join(', ') : '',
    });
    setShowForm(true);
  };

  const handleCancel = () => {
    setShowForm(false);
    setEditingArtist(null);
    setFormData({ name: '', biography: '', genres: '' });
  };

  if (loading) {
    return <div className="container">Učitavanje...</div>;
  }

  return (
    <div className="container">
      <div className="card">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <h2>Izvođači</h2>
          {isAdmin() && (
            <button className="btn btn-primary" onClick={() => setShowForm(!showForm)}>
              {showForm ? 'Otkaži' : 'Dodaj izvođača'}
            </button>
          )}
        </div>

        {showForm && isAdmin() && (
          <form onSubmit={handleSubmit} style={{ marginTop: '20px' }}>
            <div className="form-group">
              <label>Ime:</label>
              <input
                type="text"
                name="name"
                value={formData.name}
                onChange={handleChange}
                required
              />
            </div>
            <div className="form-group">
              <label>Biografija:</label>
              <textarea
                name="biography"
                value={formData.biography}
                onChange={handleChange}
              />
            </div>
            <div className="form-group">
              <label>Žanrovi (odvojeni zarezom):</label>
              <input
                type="text"
                name="genres"
                value={formData.genres}
                onChange={handleChange}
                placeholder="Rock, Pop, Jazz"
              />
            </div>
            {error && <div className="error">{error}</div>}
            <div style={{ display: 'flex', gap: '10px' }}>
              <button type="submit" className="btn btn-primary">
                {editingArtist ? 'Ažuriraj' : 'Dodaj'}
              </button>
              <button type="button" className="btn btn-secondary" onClick={handleCancel}>
                Otkaži
              </button>
            </div>
          </form>
        )}

        {error && !showForm && <div className="error">{error}</div>}

        <div style={{ marginTop: '20px' }}>
          {artists.length === 0 ? (
            <p>Nema izvođača.</p>
          ) : (
            artists.map((artist) => (
              <div
                key={artist.id}
                className="list-item"
                onClick={() => navigate(`/artists/${artist.id}`)}
              >
                <h3>{artist.name}</h3>
                {artist.biography && <p>{artist.biography}</p>}
                {artist.genres && artist.genres.length > 0 && (
                  <div style={{ marginTop: '10px' }}>
                    {artist.genres.map((genre, idx) => (
                      <span key={idx} className="genre-tag">{genre}</span>
                    ))}
                  </div>
                )}
                {isAdmin() && (
                  <button
                    className="btn btn-secondary"
                    style={{ marginTop: '10px' }}
                    onClick={(e) => {
                      e.stopPropagation();
                      handleEdit(artist);
                    }}
                  >
                    Izmeni
                  </button>
                )}
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
};

export default Artists;
