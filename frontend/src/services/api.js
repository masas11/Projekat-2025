const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8081';

// Za HTTPS development, postavite REACT_APP_API_URL=https://localhost:8081

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
        const error = data.error || data.message || errorMessage || `HTTP error! status: ${response.status}`;
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

  async deleteSong(id) {
    return this.request(`/api/content/songs/${id}`, {
      method: 'DELETE',
    });
  }

  getStreamUrl(songId) {
    return `${this.baseURL}/api/content/songs/${songId}/stream`;
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
    return this.request(`/api/subscriptions/subscribe-genre?genre=${genre}&userId=${userId}`, {
      method: 'POST',
    });
  }

  async unsubscribeFromGenre(genre, userId) {
    return this.request(`/api/subscriptions/subscribe-genre?genre=${genre}&userId=${userId}`, {
      method: 'DELETE',
    });
  }
}

export default new ApiService();
