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
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedGenre, setSelectedGenre] = useState('');
  const [formData, setFormData] = useState({
    name: '',
    biography: '',
    genres: '',
  });
  const { isAdmin, user, isAuthenticated } = useAuth();
  const navigate = useNavigate();
  const [subscriptionMessage, setSubscriptionMessage] = useState('');
  const [isSubscribing, setIsSubscribing] = useState(false);
  const [subscribedGenres, setSubscribedGenres] = useState([]);

  useEffect(() => {
    loadArtists();
    if (isAuthenticated && user) {
      loadSubscriptions();
    }
  }, [isAuthenticated, user]);

  useEffect(() => {
    filterArtists();
  }, [artists, searchTerm, selectedGenre]);

  const filterArtists = () => {
    let filtered = artists;

    // Filter by search term
    if (searchTerm) {
      filtered = filtered.filter(artist =>
        artist.name.toLowerCase().includes(searchTerm.toLowerCase())
      );
    }

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
    return (
      <div className="container" style={{ paddingTop: '60px', textAlign: 'center' }}>
        <div style={{ fontSize: '48px', marginBottom: '20px', animation: 'spin 1s linear infinite' }}>⏳</div>
        <p style={{ fontSize: '18px', color: '#666' }}>Učitavanje izvođača...</p>
      </div>
    );
  }

  // Get all unique genres from artists
  const allGenres = [...new Set(artists.flatMap(artist => artist.genres || []))].sort();

  return (
    <div style={{ minHeight: 'calc(100vh - 80px)', paddingTop: '40px', paddingBottom: '40px' }}>
      <div className="container">
        {/* Header Section */}
        <div style={{
          background: 'linear-gradient(135deg, rgba(255,255,255,0.95) 0%, rgba(255,255,255,0.9) 100%)',
          backdropFilter: 'blur(10px)',
          borderRadius: '20px',
          padding: '40px',
          marginBottom: '30px',
          boxShadow: '0 20px 60px rgba(0,0,0,0.15)',
          border: '1px solid rgba(255,255,255,0.3)'
        }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '20px' }}>
            <div>
              <div style={{ display: 'flex', alignItems: 'center', gap: '16px', marginBottom: '8px' }}>
                <div style={{
                  width: '60px',
                  height: '60px',
                  borderRadius: '16px',
                  background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: '32px',
                  boxShadow: '0 8px 20px rgba(102, 126, 234, 0.3)'
                }}>
                  🎤
                </div>
                <div>
                  <h1 style={{ 
                    margin: 0, 
                    fontSize: '36px', 
                    fontWeight: '700',
                    background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                    WebkitBackgroundClip: 'text',
                    WebkitTextFillColor: 'transparent',
                    backgroundClip: 'text'
                  }}>
                    Izvođači
                  </h1>
                  <p style={{ margin: '4px 0 0 0', color: '#666', fontSize: '14px' }}>
                    Otkrijte svoje omiljene umetnike
                  </p>
                </div>
              </div>
            </div>
            {isAdmin() && (
              <button 
                className="btn btn-primary" 
                onClick={() => setShowForm(!showForm)}
                style={{
                  padding: '14px 28px',
                  fontSize: '16px',
                  fontWeight: '600',
                  borderRadius: '12px',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '8px'
                }}
              >
                {showForm ? '✕ Otkaži' : '➕ Dodaj izvođača'}
              </button>
            )}
          </div>

          {/* Add/Edit Form */}
          {showForm && isAdmin() && (
            <div style={{
              marginTop: '30px',
              padding: '30px',
              background: 'linear-gradient(135deg, rgba(102, 126, 234, 0.05) 0%, rgba(118, 75, 162, 0.05) 100%)',
              borderRadius: '16px',
              border: '2px solid rgba(102, 126, 234, 0.2)'
            }}>
              <h3 style={{ marginTop: 0, marginBottom: '24px', fontSize: '24px', fontWeight: '600', color: '#333' }}>
                {editingArtist ? '✏️ Izmeni izvođača' : '➕ Dodaj novog izvođača'}
              </h3>
              <form onSubmit={handleSubmit}>
                <div className="form-group">
                  <label>🎭 Ime izvođača:</label>
                  <input
                    type="text"
                    name="name"
                    value={formData.name}
                    onChange={handleChange}
                    required
                    placeholder="Unesite ime izvođača"
                  />
                </div>
                <div className="form-group">
                  <label>📖 Biografija:</label>
                  <textarea
                    name="biography"
                    value={formData.biography}
                    onChange={handleChange}
                    placeholder="Unesite biografiju izvođača..."
                    rows="4"
                  />
                </div>
                <div className="form-group">
                  <label>🎵 Žanrovi (odvojeni zarezom):</label>
                  <input
                    type="text"
                    name="genres"
                    value={formData.genres}
                    onChange={handleChange}
                    placeholder="Rock, Pop, Jazz"
                  />
                  <small style={{ display: 'block', marginTop: '6px', color: '#666', fontSize: '12px' }}>
                    Primer: Pop, Rock, Jazz, Electronic
                  </small>
                </div>
                {error && (
                  <div className="error" style={{ marginBottom: '16px' }}>
                    <span>⚠️</span>
                    <span>{error}</span>
                  </div>
                )}
                <div style={{ display: 'flex', gap: '12px', marginTop: '20px' }}>
                  <button type="submit" className="btn btn-primary" style={{ flex: 1 }}>
                    {editingArtist ? '💾 Sačuvaj izmene' : '➕ Dodaj izvođača'}
                  </button>
                  <button type="button" className="btn btn-secondary" onClick={handleCancel}>
                    ✕ Otkaži
                  </button>
                </div>
              </form>
            </div>
          )}

          {error && !showForm && (
            <div className="error" style={{ marginTop: '20px' }}>
              <span>⚠️</span>
              <span>{error}</span>
            </div>
          )}

          {/* Search and Filter Controls */}
          <div style={{ 
            marginTop: '30px', 
            padding: '24px', 
            background: 'linear-gradient(135deg, rgba(102, 126, 234, 0.08) 0%, rgba(118, 75, 162, 0.08) 100%)',
            borderRadius: '16px',
            border: '1px solid rgba(102, 126, 234, 0.2)'
          }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginBottom: '20px' }}>
              <span style={{ fontSize: '24px' }}>🔍</span>
              <h3 style={{ margin: 0, fontSize: '20px', fontWeight: '600', color: '#333' }}>Pretraga i Filtriranje</h3>
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))', gap: '16px', marginBottom: '16px' }}>
              <div className="form-group" style={{ marginBottom: 0 }}>
                <label style={{ marginBottom: '8px' }}>🔎 Pretraga po nazivu:</label>
                <input
                  type="text"
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  placeholder="Unesite naziv umetnika..."
                />
              </div>
              <div className="form-group" style={{ marginBottom: 0 }}>
                <label style={{ marginBottom: '8px' }}>🎵 Filtriranje po žanru:</label>
                <div style={{ display: 'flex', gap: '8px' }}>
                  <select
                    value={selectedGenre}
                    onChange={(e) => setSelectedGenre(e.target.value)}
                    style={{ flex: 1 }}
                  >
                    <option value="">🎵 Svi žanrovi</option>
                    {allGenres.map(genre => (
                      <option key={genre} value={genre}>{genre}</option>
                    ))}
                  </select>
                  {isAuthenticated && selectedGenre && (
                    <button
                      className={subscribedGenres.includes(selectedGenre) ? "btn btn-secondary" : "btn btn-primary"}
                      onClick={() => handleSubscribeToGenre(selectedGenre)}
                      disabled={isSubscribing}
                      style={{ 
                        whiteSpace: 'nowrap',
                        padding: '14px 20px',
                        fontSize: '16px'
                      }}
                      title={subscribedGenres.includes(selectedGenre) ? `Odjavite se sa žanra: ${selectedGenre}` : `Pretplati se na žanr: ${selectedGenre}`}
                    >
                      {isSubscribing 
                        ? '⏳' 
                        : (subscribedGenres.includes(selectedGenre) ? '✓ Pretplaćen' : '🔔 Pretplati se')}
                    </button>
                  )}
                </div>
              </div>
            </div>
            {subscriptionMessage && (
              <div className="success" style={{ marginTop: '12px', marginBottom: '12px' }}>
                <span>✓</span>
                <span>{subscriptionMessage}</span>
              </div>
            )}
            <div style={{ 
              display: 'flex', 
              alignItems: 'center', 
              gap: '8px',
              padding: '12px',
              background: 'rgba(102, 126, 234, 0.1)',
              borderRadius: '8px',
              fontSize: '14px',
              fontWeight: '500',
              color: '#667eea'
            }}>
              <span>📊</span>
              <span>Pronađeno umetnika: <strong>{filteredArtists.length}</strong> od <strong>{artists.length}</strong></span>
            </div>
          </div>
        </div>

        {/* Artists Grid */}
        <div style={{ marginTop: '30px' }}>
          {filteredArtists.length === 0 ? (
            <div style={{
              background: 'linear-gradient(135deg, rgba(255,255,255,0.95) 0%, rgba(255,255,255,0.9) 100%)',
              backdropFilter: 'blur(10px)',
              borderRadius: '20px',
              padding: '60px 40px',
              textAlign: 'center',
              boxShadow: '0 20px 60px rgba(0,0,0,0.15)',
              border: '1px solid rgba(255,255,255,0.3)'
            }}>
              <div style={{ fontSize: '64px', marginBottom: '20px' }}>🎭</div>
              <h3 style={{ fontSize: '24px', fontWeight: '600', color: '#333', marginBottom: '12px' }}>
                {artists.length === 0 ? 'Nema izvođača' : 'Nema umetnika koji odgovaraju filtru'}
              </h3>
              <p style={{ color: '#666', fontSize: '16px' }}>
                {artists.length === 0 
                  ? 'Dodajte prvog izvođača da biste počeli!' 
                  : 'Pokušajte da promenite filter da biste videli više rezultata.'}
              </p>
            </div>
          ) : (
            <div style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))',
              gap: '24px'
            }}>
              {filteredArtists.map((artist) => (
                <div
                  key={artist.id}
                  style={{
                    background: 'linear-gradient(135deg, rgba(255,255,255,0.95) 0%, rgba(255,255,255,0.9) 100%)',
                    backdropFilter: 'blur(10px)',
                    borderRadius: '20px',
                    padding: '28px',
                    boxShadow: '0 10px 30px rgba(0,0,0,0.1)',
                    border: '1px solid rgba(255,255,255,0.3)',
                    cursor: 'pointer',
                    transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                    position: 'relative',
                    overflow: 'hidden'
                  }}
                  onMouseEnter={(e) => {
                    e.currentTarget.style.transform = 'translateY(-8px)';
                    e.currentTarget.style.boxShadow = '0 20px 50px rgba(102, 126, 234, 0.25)';
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.transform = 'translateY(0)';
                    e.currentTarget.style.boxShadow = '0 10px 30px rgba(0,0,0,0.1)';
                  }}
                  onClick={() => navigate(`/artists/${artist.id}`)}
                >
                  {/* Artist Avatar */}
                  <div style={{
                    width: '80px',
                    height: '80px',
                    borderRadius: '20px',
                    background: `linear-gradient(135deg, 
                      hsl(${(artist.id.charCodeAt(0) * 137.508) % 360}, 70%, 60%) 0%, 
                      hsl(${(artist.id.charCodeAt(0) * 137.508 + 60) % 360}, 70%, 50%) 100%)`,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontSize: '40px',
                    marginBottom: '20px',
                    boxShadow: '0 8px 20px rgba(0,0,0,0.15)',
                    transition: 'transform 0.3s ease'
                  }}
                  onMouseEnter={(e) => {
                    e.currentTarget.style.transform = 'scale(1.1) rotate(5deg)';
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.transform = 'scale(1) rotate(0deg)';
                  }}
                  >
                    🎤
                  </div>

                  {/* Artist Name */}
                  <h3 style={{
                    margin: '0 0 12px 0',
                    fontSize: '24px',
                    fontWeight: '700',
                    color: '#333',
                    lineHeight: '1.3'
                  }}>
                    {artist.name}
                  </h3>

                  {/* Biography */}
                  {artist.biography && (
                    <p style={{
                      margin: '0 0 16px 0',
                      color: '#666',
                      fontSize: '14px',
                      lineHeight: '1.6',
                      display: '-webkit-box',
                      WebkitLineClamp: 3,
                      WebkitBoxOrient: 'vertical',
                      overflow: 'hidden',
                      textOverflow: 'ellipsis'
                    }}>
                      {artist.biography}
                    </p>
                  )}

                  {/* Genres */}
                  {artist.genres && artist.genres.length > 0 && (
                    <div style={{ 
                      marginTop: '16px',
                      display: 'flex',
                      flexWrap: 'wrap',
                      gap: '8px'
                    }}>
                      {artist.genres.map((genre, idx) => (
                        <span 
                          key={idx} 
                          className="genre-tag"
                          style={{
                            padding: '6px 14px',
                            background: 'linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%)',
                            color: '#667eea',
                            border: '1px solid rgba(102, 126, 234, 0.2)',
                            fontWeight: '500',
                            fontSize: '12px'
                          }}
                        >
                          {genre}
                        </span>
                      ))}
                    </div>
                  )}

                  {/* Admin Actions */}
                  {isAdmin() && (
                    <div style={{ 
                      display: 'flex', 
                      gap: '8px', 
                      marginTop: '20px',
                      paddingTop: '20px',
                      borderTop: '1px solid #e0e0e0'
                    }}>
                      <button
                        className="btn btn-secondary"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleEdit(artist);
                        }}
                        style={{
                          flex: 1,
                          padding: '10px 16px',
                          fontSize: '14px',
                          fontWeight: '500'
                        }}
                      >
                        ✏️ Izmeni
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
                        style={{
                          flex: 1,
                          padding: '10px 16px',
                          fontSize: '14px',
                          fontWeight: '500'
                        }}
                      >
                        🗑️ Obriši
                      </button>
                    </div>
                  )}

                  {/* View Arrow */}
                  <div style={{
                    position: 'absolute',
                    top: '24px',
                    right: '24px',
                    fontSize: '20px',
                    opacity: 0.3,
                    transition: 'all 0.3s ease'
                  }}
                  onMouseEnter={(e) => {
                    e.currentTarget.style.opacity = 1;
                    e.currentTarget.style.transform = 'translateX(4px)';
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.opacity = 0.3;
                    e.currentTarget.style.transform = 'translateX(0)';
                  }}
                  >
                    →
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default Artists;
