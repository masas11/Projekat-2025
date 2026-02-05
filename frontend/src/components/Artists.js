import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import api from '../services/api';

const Artists = () => {
  const [artists, setArtists] = useState([]);
  const [filteredArtists, setFilteredArtists] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showForm, setShowForm] = useState(false);
  const [editingArtist, setEditingArtist] = useState(null);
  const [selectedGenre, setSelectedGenre] = useState('');
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

  useEffect(() => {
    filterArtists();
  }, [artists, selectedGenre]);

  const filterArtists = () => {
    let filtered = artists;

    // Filter by genre
    if (selectedGenre) {
      filtered = filtered.filter(artist =>
        artist.genres && artist.genres.includes(selectedGenre)
      );
    }

    setFilteredArtists(filtered);
  };

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

        {/* Genre Filter */}
        <div style={{ 
          marginTop: '20px', 
          padding: '15px', 
          backgroundColor: '#f8f9fa', 
          borderRadius: '5px',
          border: '1px solid #ddd'
        }}>
          <h4>Filtriranje po žanru</h4>
          <div style={{ flex: 1 }}>
            <label>Izaberite žanr:</label>
            <select
              value={selectedGenre}
              onChange={(e) => setSelectedGenre(e.target.value)}
              style={{ 
                width: '100%', 
                padding: '8px', 
                border: '1px solid #ddd',
                borderRadius: '4px'
              }}
            >
              <option value="">Svi žanrovi</option>
              <option value="Pop">Pop</option>
              <option value="Rock">Rock</option>
              <option value="Jazz">Jazz</option>
              <option value="Classical">Classical</option>
              <option value="Electronic">Electronic</option>
              <option value="Hip-Hop">Hip-Hop</option>
              <option value="Country">Country</option>
              <option value="R&B">R&B</option>
              <option value="Reggae">Reggae</option>
            </select>
          </div>
          <div style={{ fontSize: '0.9em', color: '#666', marginTop: '10px' }}>
            Pronađeno umetnika: {filteredArtists.length} od {artists.length}
          </div>
        </div>

        <div style={{ marginTop: '20px' }}>
          {filteredArtists.length === 0 ? (
            <p>{artists.length === 0 ? 'Nema izvođača.' : 'Nema umetnika koji odgovaraju filtru.'}</p>
          ) : (
            filteredArtists.map((artist) => (
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
                  <div style={{ display: 'flex', gap: '10px', marginTop: '10px' }}>
                    <button
                      className="btn btn-secondary"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleEdit(artist);
                      }}
                    >
                      Izmeni
                    </button>
                    <button
                      className="btn btn-danger"
                      onClick={async (e) => {
                        e.stopPropagation();
                        if (window.confirm(`Da li ste sigurni da želite da obrišete izvođača "${artist.name}"?`)) {
                          try {
                            await api.deleteArtist(artist.id);
                            loadArtists();
                          } catch (err) {
                            setError(err.message || 'Greška pri brisanju izvođača');
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

export default Artists;
