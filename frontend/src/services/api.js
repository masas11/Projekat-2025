const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8081';

// HTTPS je omogućen za komunikaciju sa API Gateway-em
// Za development sa self-signed sertifikatima, browser će tražiti potvrdu

class ApiService {
  constructor() {
    this.baseURL = API_BASE_URL;
  }

  async request(endpoint, options = {}) {
    const url = `${this.baseURL}${endpoint}`;
    const config = {
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      ...options,
    };

    // Add auth token if available
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    } else {
      console.warn('No token found in localStorage for request to:', endpoint);
    }

    try {
      const response = await fetch(url, config);
      
      // Handle empty responses
      const contentType = response.headers.get('content-type');
      let data = {};
      let errorMessage = '';
      
      if (contentType && contentType.includes('application/json')) {
        const text = await response.text();
        data = text ? JSON.parse(text) : {};
      } else {
        // If not JSON, read as text (for error messages)
        const text = await response.text();
        if (text) {
          errorMessage = text;
        }
      }
      
      if (!response.ok) {
        // Try to extract error message from various formats
        let error = data.error || data.message || errorMessage;
        
        // If no message in body, use status-based message
        if (!error) {
          if (response.status === 409) {
            error = 'Korisnik sa ovim korisničkim imenom ili email adresom već postoji.';
          } else if (response.status === 400) {
            error = 'Neispravni podaci. Proverite unete podatke.';
          } else if (response.status === 401) {
            error = 'Neautorizovan pristup.';
          } else if (response.status === 403) {
            error = 'Pristup zabranjen.';
          } else {
            error = `HTTP greška! Status: ${response.status}`;
          }
        }
        
        throw new Error(error);
      }
      
      return data;
    } catch (error) {
      throw error;
    }
  }

  // Users Service
  async register(userData) {
    return this.request('/api/users/register', {
      method: 'POST',
      body: JSON.stringify(userData),
    });
  }

  async requestOTP(credentials) {
    return this.request('/api/users/login/request-otp', {
      method: 'POST',
      body: JSON.stringify(credentials),
    });
  }

  async verifyOTP(username, otp) {
    return this.request('/api/users/login/verify-otp', {
      method: 'POST',
      body: JSON.stringify({ username, otp }),
    });
  }

  async changePassword(data) {
    return this.request('/api/users/password/change', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async requestPasswordReset(email) {
    return this.request('/api/users/password/reset/request', {
      method: 'POST',
      body: JSON.stringify({ email }),
    });
  }

  async resetPassword(token, newPassword) {
    return this.request('/api/users/password/reset', {
      method: 'POST',
      body: JSON.stringify({ token, newPassword }),
    });
  }

  async verifyEmail(token) {
    // Token is already URL decoded when read from searchParams
    // Use URLSearchParams to properly encode the token without double-encoding
    const params = new URLSearchParams({ token });
    return this.request(`/api/users/verify-email?${params.toString()}`);
  }

  async requestMagicLink(email) {
    return this.request('/api/users/recover/request', {
      method: 'POST',
      body: JSON.stringify({ email }),
    });
  }

  async verifyMagicLink(token) {
    // Use URLSearchParams to properly encode the token
    const params = new URLSearchParams({ token });
    return this.request(`/api/users/recover/verify?${params.toString()}`);
  }

  async logout() {
    return this.request('/api/users/logout', {
      method: 'POST',
    });
  }

  // Content Service - Artists
  async getArtists() {
    return this.request('/api/content/artists');
  }

  async getArtist(id) {
    return this.request(`/api/content/artists/${id}`);
  }

  async createArtist(artistData) {
    return this.request('/api/content/artists', {
      method: 'POST',
      body: JSON.stringify(artistData),
    });
  }

  async updateArtist(id, artistData) {
    return this.request(`/api/content/artists/${id}`, {
      method: 'PUT',
      body: JSON.stringify(artistData),
    });
  }

  async deleteArtist(id) {
    return this.request(`/api/content/artists/${id}`, {
      method: 'DELETE',
    });
  }

  // Content Service - Albums
  async getAlbums() {
    return this.request('/api/content/albums');
  }

  async getAlbum(id) {
    return this.request(`/api/content/albums/${id}`);
  }

  async getAlbumsByArtist(artistId) {
    return this.request(`/api/content/albums/by-artist?artistId=${artistId}`);
  }

  async createAlbum(albumData) {
    return this.request('/api/content/albums', {
      method: 'POST',
      body: JSON.stringify(albumData),
    });
  }

  async updateAlbum(id, albumData) {
    return this.request(`/api/content/albums/${id}`, {
      method: 'PUT',
      body: JSON.stringify(albumData),
    });
  }

  async deleteAlbum(id) {
    return this.request(`/api/content/albums/${id}`, {
      method: 'DELETE',
    });
  }

  // Content Service - Songs
  async getSongs() {
    return this.request('/api/content/songs');
  }

  async getSong(id) {
    return this.request(`/api/content/songs/${id}`);
  }

  async getSongsByAlbum(albumId) {
    return this.request(`/api/content/songs/by-album?albumId=${albumId}`);
  }

  async createSong(songData) {
    return this.request('/api/content/songs', {
      method: 'POST',
      body: JSON.stringify(songData),
    });
  }

  async updateSong(id, songData) {
    return this.request(`/api/content/songs/${id}`, {
      method: 'PUT',
      body: JSON.stringify(songData),
    });
  }

  async uploadAudioFile(songId, audioFile) {
    const formData = new FormData();
    formData.append('audio', audioFile);
    formData.append('songId', songId);

    const token = localStorage.getItem('token');
    if (!token) {
      throw new Error('Not authenticated');
    }

    const response = await fetch(`${this.baseURL}/api/content/songs/${songId}/upload`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
      },
      body: formData,
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Upload failed' }));
      throw new Error(error.error || error.message || 'Upload failed');
    }

    return response.json();
  }

  async deleteSong(id) {
    return this.request(`/api/content/songs/${id}`, {
      method: 'DELETE',
    });
  }

  getStreamUrl(songId) {
    const token = localStorage.getItem('token');
    // Don't add timestamp here - it will be added by AudioPlayer when needed
    if (token) {
      return `${this.baseURL}/api/content/songs/${songId}/stream?token=${encodeURIComponent(token)}`;
    }
    return `${this.baseURL}/api/content/songs/${songId}/stream`;
  }

  // Get most played songs (2.12)
  async getMostPlayedSongs(limit = 10) {
    return this.request(`/api/content/songs/most-played?limit=${limit}`);
  }

  // Analytics Service (1.15)
  // Direktno pozivamo analytics-service jer API Gateway trenutno ne prosleđuje podatke ispravno
  async getUserActivities(limit = 50, type = null, userId) {
    const params = new URLSearchParams();
    if (limit) params.append('limit', limit);
    if (type) params.append('type', type);
    if (userId) params.append('userId', userId);

    const base = process.env.REACT_APP_ANALYTICS_URL || 'http://localhost:8007';
    const url = `${base}/activities?${params.toString()}`;

    console.log('getUserActivities called (direct analytics) with:', { limit, type, userId, url });

    const token = localStorage.getItem('token');
    const headers = {
      'Content-Type': 'application/json',
    };
    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }

    const response = await fetch(url, { headers });
    const text = await response.text();
    if (!response.ok) {
      throw new Error(text || `Greška prilikom učitavanja aktivnosti (status ${response.status})`);
    }

    try {
      return text ? JSON.parse(text) : [];
    } catch (e) {
      console.error('Greška pri parsiranju aktivnosti:', e, 'raw:', text);
      throw new Error('Neispravan odgovor servera za aktivnosti');
    }
  }

  // Get user analytics (1.16)
  async getUserAnalytics(userId) {
    return this.request(`/api/analytics/analytics?userId=${userId}`);
  }

  // Notifications Service
  // userId is automatically extracted from JWT token by API Gateway
  async getNotifications() {
    return this.request('/api/notifications');
  }

  // Subscriptions Service
  async getSubscriptions() {
    try {
      const result = await this.request('/api/subscriptions');
      return Array.isArray(result) ? result : [];
    } catch (err) {
      console.error('Error getting subscriptions:', err);
      return [];
    }
  }

  async subscribeToArtist(artistId, userId) {
    return this.request(`/api/subscriptions/subscribe-artist?artistId=${artistId}&userId=${userId}`, {
      method: 'POST',
    });
  }

  async unsubscribeFromArtist(artistId, userId) {
    return this.request(`/api/subscriptions/subscribe-artist?artistId=${artistId}&userId=${userId}`, {
      method: 'DELETE',
    });
  }

  async subscribeToGenre(genre, userId) {
    const encodedGenre = encodeURIComponent(genre);
    return this.request(`/api/subscriptions/subscribe-genre?genre=${encodedGenre}&userId=${userId}`, {
      method: 'POST',
    });
  }

  async unsubscribeFromGenre(genre, userId) {
    const encodedGenre = encodeURIComponent(genre);
    return this.request(`/api/subscriptions/subscribe-genre?genre=${encodedGenre}&userId=${userId}`, {
      method: 'DELETE',
    });
  }

  // Ratings Service
  async rateSong(songId, rating, userId) {
    return this.request(`/api/ratings/rate-song?songId=${songId}&rating=${rating}&userId=${userId}`, {
      method: 'POST',
    });
  }

  async getRating(songId, userId) {
    return this.request(`/api/ratings/get-rating?songId=${songId}&userId=${userId}`);
  }

  async deleteRating(songId, userId) {
    return this.request(`/api/ratings/delete-rating?songId=${songId}&userId=${userId}`, {
      method: 'DELETE',
    });
  }
}

export default new ApiService();
