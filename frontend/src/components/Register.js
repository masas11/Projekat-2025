import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../services/api';

const Register = () => {
  const [formData, setFormData] = useState({
    firstName: '',
    lastName: '',
    email: '',
    username: '',
    password: '',
    confirmPassword: '',
  });
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleChange = (e) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    });
  };

  const validatePassword = (password) => {
    if (password.length < 8) {
      return 'Lozinka mora imati najmanje 8 karaktera';
    }
    if (!/[A-Z]/.test(password)) {
      return 'Lozinka mora sadržati najmanje jedno veliko slovo';
    }
    if (!/[0-9]/.test(password)) {
      return 'Lozinka mora sadržati najmanje jedan broj';
    }
    return null;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    if (formData.password !== formData.confirmPassword) {
      setError('Lozinke se ne poklapaju');
      return;
    }

    const passwordError = validatePassword(formData.password);
    if (passwordError) {
      setError(passwordError);
      return;
    }

    setLoading(true);

    try {
      await api.register(formData);
      
      // Success - show message and redirect to login
      setSuccess('Uspešna registracija! Email za verifikaciju je poslat. Proverite MailHog (http://localhost:8025) i kliknite na link za verifikaciju.');
      
      // Redirect to login after 3 seconds
      setTimeout(() => {
        navigate('/login', { 
          state: { 
            message: 'Registracija uspešna! Proverite MailHog (http://localhost:8025) za verifikacioni email.' 
          } 
        });
      }, 3000);
    } catch (err) {
      // Better error handling for 409 Conflict (user already exists)
      if (err.message && (err.message.includes('409') || err.message.includes('already exists') || err.message.includes('user already exists'))) {
        setError('Korisnik sa ovim korisničkim imenom ili email adresom već postoji. Molimo koristite drugačije podatke ili se prijavite.');
      } else {
        setError(err.message || 'Greška pri registraciji');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ 
      minHeight: 'calc(100vh - 80px)',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      padding: '40px 20px',
      background: 'linear-gradient(135deg, rgba(102, 126, 234, 0.05) 0%, rgba(118, 75, 162, 0.05) 100%)'
    }}>
      <div className="card" style={{ 
        maxWidth: '520px', 
        width: '100%',
        background: 'linear-gradient(135deg, rgba(255,255,255,0.95) 0%, rgba(255,255,255,0.9) 100%)',
        backdropFilter: 'blur(10px)',
        boxShadow: '0 20px 60px rgba(0,0,0,0.15)',
        border: '1px solid rgba(255,255,255,0.3)'
      }}>
        <div style={{ textAlign: 'center', marginBottom: '40px' }}>
          <div style={{
            width: '80px',
            height: '80px',
            borderRadius: '20px',
            background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontSize: '40px',
            margin: '0 auto 20px',
            boxShadow: '0 8px 20px rgba(102, 126, 234, 0.3)'
          }}>
            ✨
          </div>
          <h2 style={{ 
            fontSize: '32px', 
            fontWeight: '700',
            marginBottom: '8px',
            background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
            WebkitBackgroundClip: 'text',
            WebkitTextFillColor: 'transparent',
            backgroundClip: 'text'
          }}>
            Registruj se
          </h2>
          <p style={{ color: '#666', fontSize: '14px' }}>Kreirajte svoj nalog i počnite slušati</p>
        </div>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Ime:</label>
            <input
              type="text"
              name="firstName"
              value={formData.firstName}
              onChange={handleChange}
              required
            />
          </div>
          <div className="form-group">
            <label>Prezime:</label>
            <input
              type="text"
              name="lastName"
              value={formData.lastName}
              onChange={handleChange}
              required
            />
          </div>
          <div className="form-group">
            <label>Email:</label>
            <input
              type="email"
              name="email"
              value={formData.email}
              onChange={handleChange}
              required
            />
          </div>
          <div className="form-group">
            <label>Korisničko ime:</label>
            <input
              type="text"
              name="username"
              value={formData.username}
              onChange={handleChange}
              required
            />
          </div>
          <div className="form-group">
            <label>Lozinka:</label>
            <input
              type="password"
              name="password"
              value={formData.password}
              onChange={handleChange}
              required
            />
            <small style={{ color: '#666', fontSize: '12px', display: 'block', marginTop: '5px' }}>
              Lozinka mora imati najmanje 8 karaktera, jedno veliko slovo i jedan broj
            </small>
          </div>
          <div className="form-group">
            <label>Potvrdi lozinku:</label>
            <input
              type="password"
              name="confirmPassword"
              value={formData.confirmPassword}
              onChange={handleChange}
              required
            />
          </div>
          {error && (
            <div className="error">
              <span>⚠️</span>
              <span>{error}</span>
            </div>
          )}
          {success && (
            <div className="success">
              <span>✅</span>
              <div>
                <p style={{ marginBottom: '10px', fontWeight: '500' }}>{success}</p>
                <p style={{ fontSize: '13px', color: '#666', marginTop: '8px' }}>
                  Preusmeravanje na stranicu za prijavu za 3 sekunde...
                </p>
                <div style={{ marginTop: '15px', padding: '12px', backgroundColor: '#e3f2fd', borderRadius: '8px', border: '1px solid #90caf9' }}>
                  <p style={{ fontSize: '13px', marginBottom: '8px', fontWeight: '500' }}>📧 MailHog Web UI:</p>
                  <a 
                    href="http://localhost:8025" 
                    target="_blank" 
                    rel="noopener noreferrer"
                    style={{ 
                      color: '#1976d2', 
                      textDecoration: 'underline',
                      fontSize: '14px',
                      fontWeight: '500'
                    }}
                  >
                    http://localhost:8025
                  </a>
                </div>
              </div>
            </div>
          )}
          <button type="submit" className="btn btn-primary" disabled={loading} style={{ width: '100%', marginTop: '10px' }}>
            {loading ? 'Registracija...' : 'Registruj se'}
          </button>
        </form>
        <div style={{ marginTop: '20px', textAlign: 'center', fontSize: '14px', color: '#666' }}>
          Već imate nalog? <a href="/login" style={{ color: '#667eea', textDecoration: 'none', fontWeight: '500' }}>Prijavite se</a>
        </div>
      </div>
    </div>
  );
};

export default Register;
