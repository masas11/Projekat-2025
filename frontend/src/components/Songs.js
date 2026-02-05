import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import api from '../services/api';

const Songs = () => {
  const [songs, setSongs] = useState([]);
  const [filteredSongs, setFilteredSongs] = useState([]);
  const [albums, setAlbums] = useState([]);
  const [artists, setArtists] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showForm, setShowForm] = useState(false);
  const [editingSong, setEditingSong] = useState(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedGenre, setSelectedGenre] = useState('');
  
  // Predefined genres for consistency
  const predefinedGenres = [
    'Pop', 'Rock', 'Jazz', 'Classical', 'Electronic', 
    'Hip-Hop', 'Country', 'R&B', 'Reggae', 'Blues',
    'Metal', 'Folk', 'Soul', 'Funk', 'Punk'
  ];
  
  const [formData, setFormData] = useState({
    name: '',
    duration: '',
    genre: '',
    albumId: '',
    selectedArtistIds: [],
    audioFileUrl: '',
  });
  const { isAdmin } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    loadSongs();
    loadAlbums();
    loadArtists();
  }, []);

  useEffect(() => {
    filterSongs();
  }, [songs, searchTerm, selectedGenre]);

  const filterSongs = () => {
    let filtered = songs;

    // Filter by search term
    if (searchTerm) {
      filtered = filtered.filter(song =>
        song.name.toLowerCase().includes(searchTerm.toLowerCase())
      );
    }

    // Filter by genre
    if (selectedGenre) {
      filtered = filtered.filter(song =>
        song.genre === selectedGenre
      );
    }

    setFilteredSongs(filtered);
  };

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

    const songData = {
      name: formData.name,
      duration: parseInt(formData.duration),
      genre: formData.genre,
      albumId: formData.albumId,
      artistIds: formData.selectedArtistIds,
      audioFileUrl: formData.audioFileUrl,
    };

    try {
      if (editingSong) {
        await api.updateSong(editingSong.id, songData);
      } else {
        await api.createSong(songData);
      }
      setShowForm(false);
      setEditingSong(null);
      setFormData({ name: '', duration: '', genre: '', albumId: '', selectedArtistIds: [], audioFileUrl: '' });
      loadSongs();
    } catch (err) {
      setError(err.message || 'Greška pri čuvanju pesme');
    }
  };

  const handleEdit = (song) => {
    setEditingSong(song);
    setFormData({
      name: song.name,
      duration: song.duration || '',
      genre: song.genre || '',
      albumId: song.albumId || song.albumID || '',
      selectedArtistIds: song.artistIds || song.artistIDs || [],
      audioFileUrl: song.audioFileUrl || '',
    });
    setShowForm(true);
  };

  const handleCancel = () => {
    setShowForm(false);
    setEditingSong(null);
    setFormData({ name: '', duration: '', genre: '', albumId: '', selectedArtistIds: [], audioFileUrl: '' });
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
              <select
                name="genre"
                value={formData.genre}
                onChange={handleChange}
                required
                style={{ 
                  width: '100%', 
                  padding: '8px', 
                  border: '1px solid #ddd',
                  borderRadius: '4px'
                }}
              >
                <option value="">Izaberite žanr</option>
                {predefinedGenres.map((genre) => (
                  <option key={genre} value={genre}>
                    {genre}
                  </option>
                ))}
              </select>
            </div>
            <div className="form-group">
              <label>Album:</label>
              <select
                name="albumId"
                value={formData.albumId}
                onChange={handleChange}
                required
                style={{ 
                  width: '100%', 
                  padding: '8px', 
                  border: '1px solid #ddd',
                  borderRadius: '4px'
                }}
              >
                <option value="">Izaberite album</option>
                {albums.map((album) => (
                  <option key={album.id} value={album.id}>
                    {album.name}
                  </option>
                ))}
              </select>
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
            <div className="form-group">
              <label>Audio File URL:</label>
              <input
                type="text"
                name="audioFileUrl"
                value={formData.audioFileUrl}
                onChange={handleChange}
                placeholder="/music/Lady Gaga - Abracadabra.mp3 ili https://example.com/song.mp3"
                style={{ 
                  width: '100%', 
                  padding: '8px', 
                  border: '1px solid #ddd',
                  borderRadius: '4px'
                }}
              />
              <p style={{ marginTop: '5px', fontSize: '0.9em', color: '#666' }}>
                Unesite URL do audio fajla (MP3, WAV, OGG) ili lokalnu putanju<br/>
                <strong>Napomena:</strong> Koristite validne audio URL-ove za reprodukciju
              </p>
            </div>
            {error && <div className="error">{error}</div>}
            <div style={{ display: 'flex', gap: '10px' }}>
              <button type="submit" className="btn btn-primary">
                {editingSong ? 'Ažuriraj' : 'Dodaj'}
              </button>
              <button type="button" className="btn btn-secondary" onClick={handleCancel}>
                Otkaži
              </button>
            </div>
          </form>
        )}

        {error && !showForm && <div className="error">{error}</div>}

        {/* Search and Filter Controls */}
        <div style={{ 
          marginTop: '20px', 
          padding: '15px', 
          backgroundColor: '#f8f9fa', 
          borderRadius: '5px',
          border: '1px solid #ddd'
        }}>
          <h4>Pretraga i Filtriranje</h4>
          <div style={{ display: 'flex', gap: '15px', marginBottom: '10px' }}>
            <div style={{ flex: 1 }}>
              <label>Pretraga po nazivu:</label>
              <input
                type="text"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                placeholder="Unesite naziv pesme..."
                style={{ 
                  width: '100%', 
                  padding: '8px', 
                  border: '1px solid #ddd',
                  borderRadius: '4px'
                }}
              />
            </div>
            <div style={{ flex: 1 }}>
              <label>Filtriranje po žanru:</label>
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
                {predefinedGenres.map((genre) => (
                  <option key={genre} value={genre}>
                    {genre}
                  </option>
                ))}
              </select>
            </div>
          </div>
          <div style={{ fontSize: '0.9em', color: '#666' }}>
            Pronađeno pesama: {filteredSongs.length} od {songs.length}
          </div>
        </div>

        <div style={{ marginTop: '20px' }}>
          {filteredSongs.length === 0 ? (
            <p>{songs.length === 0 ? 'Nema pesama.' : 'Nema pesama koje odgovaraju pretrazi.'}</p>
          ) : (
            filteredSongs.map((song) => (
              <div
                key={song.id}
                className="list-item"
                onClick={() => navigate(`/songs/${song.id}`)}
              >
                <h3>{song.name}</h3>
                {song.duration && <p>Trajanje: {formatDuration(song.duration)}</p>}
                {song.genre && <span className="genre-tag">{song.genre}</span>}
                {isAdmin() && (
                  <div style={{ display: 'flex', gap: '10px', marginTop: '10px' }}>
                    <button
                      className="btn btn-secondary"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleEdit(song);
                      }}
                    >
                      Izmeni
                    </button>
                    <button
                      className="btn btn-danger"
                      onClick={async (e) => {
                        e.stopPropagation();
                        if (window.confirm(`Da li ste sigurni da želite da obrišete pesmu "${song.name}"?`)) {
                          try {
                            await api.deleteSong(song.id);
                            loadSongs();
                          } catch (err) {
                            setError(err.message || 'Greška pri brisanju pesme');
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

export default Songs;
