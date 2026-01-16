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
  const [editingAlbum, setEditingAlbum] = useState(null);
  const [formData, setFormData] = useState({
    name: '',
    releaseDate: '',
    genre: '',
    selectedArtistIds: [],
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
    const { name, value } = e.target;
    setFormData({
      ...formData,
      [name]: value,
    });
  };

  const handleArtistSelect = (e) => {
    const selectedOptions = Array.from(e.target.selectedOptions, option => option.value);
    setFormData({
      ...formData,
      selectedArtistIds: selectedOptions,
    });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');

    const albumData = {
      name: formData.name,
      releaseDate: formData.releaseDate ? new Date(formData.releaseDate).toISOString() : new Date().toISOString(),
      genre: formData.genre,
      artistIds: formData.selectedArtistIds,
    };

    try {
      if (editingAlbum) {
        await api.updateAlbum(editingAlbum.id, albumData);
      } else {
        await api.createAlbum(albumData);
      }
      setShowForm(false);
      setEditingAlbum(null);
      setFormData({ name: '', releaseDate: '', genre: '', selectedArtistIds: [] });
      loadAlbums();
    } catch (err) {
      setError(err.message || 'Greška pri čuvanju albuma');
    }
  };

  const handleEdit = (album) => {
    setEditingAlbum(album);
    setFormData({
      name: album.name,
      releaseDate: album.releaseDate ? new Date(album.releaseDate).toISOString().split('T')[0] : '',
      genre: album.genre || '',
      selectedArtistIds: album.artistIds || album.artistIDs || [],
    });
    setShowForm(true);
  };

  const handleCancel = () => {
    setShowForm(false);
    setEditingAlbum(null);
    setFormData({ name: '', releaseDate: '', genre: '', selectedArtistIds: [] });
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
              <label>Izvođači (držite Ctrl/Cmd za višestruki izbor):</label>
              <select
                name="selectedArtistIds"
                multiple
                value={formData.selectedArtistIds}
                onChange={handleArtistSelect}
                required
                style={{ 
                  width: '100%', 
                  padding: '8px', 
                  minHeight: '100px',
                  border: '1px solid #ddd',
                  borderRadius: '4px'
                }}
              >
                {artists.map((artist) => (
                  <option key={artist.id} value={artist.id}>
                    {artist.name}
                  </option>
                ))}
              </select>
              {formData.selectedArtistIds.length > 0 && (
                <p style={{ marginTop: '5px', fontSize: '0.9em', color: '#666' }}>
                  Izabrano: {formData.selectedArtistIds.length} izvođač(a)
                </p>
              )}
            </div>
            {error && <div className="error">{error}</div>}
            <div style={{ display: 'flex', gap: '10px' }}>
              <button type="submit" className="btn btn-primary">
                {editingAlbum ? 'Ažuriraj' : 'Dodaj'}
              </button>
              <button type="button" className="btn btn-secondary" onClick={handleCancel}>
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
                {isAdmin() && (
                  <div style={{ display: 'flex', gap: '10px', marginTop: '10px' }}>
                    <button
                      className="btn btn-secondary"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleEdit(album);
                      }}
                    >
                      Izmeni
                    </button>
                    <button
                      className="btn btn-danger"
                      onClick={async (e) => {
                        e.stopPropagation();
                        if (window.confirm(`Da li ste sigurni da želite da obrišete album "${album.name}"?`)) {
                          try {
                            await api.deleteAlbum(album.id);
                            loadAlbums();
                          } catch (err) {
                            setError(err.message || 'Greška pri brisanju albuma');
                          }
                        }
                      }}
                    >
                      Obriši
                    </button>
                  </div>
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
