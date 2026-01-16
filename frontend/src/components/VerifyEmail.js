import React, { useState, useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import api from '../services/api';

const VerifyEmail = () => {
  const [searchParams] = useSearchParams();
  const [status, setStatus] = useState('verifying'); // verifying, success, error
  const [message, setMessage] = useState('');
  const navigate = useNavigate();

  useEffect(() => {
    const token = searchParams.get('token');
    const rawUrl = window.location.href;
    
    console.log('VerifyEmail - Raw URL:', rawUrl);
    console.log('VerifyEmail - Token from URL (decoded):', token);
    console.log('VerifyEmail - Token length:', token?.length);
    
    if (!token) {
      setStatus('error');
      setMessage('Token za verifikaciju nije pronađen u URL-u. Proverite link.');
      return;
    }

    const verifyEmail = async () => {
      try {
        console.log('VerifyEmail - Calling API with token:', token);
        const apiUrl = `/api/users/verify-email?token=${encodeURIComponent(token)}`;
        console.log('VerifyEmail - API URL:', apiUrl);
        await api.verifyEmail(token);
        setStatus('success');
        setMessage('Email je uspešno verifikovan! Sada se možete prijaviti.');
        setTimeout(() => {
          navigate('/login');
        }, 3000);
      } catch (err) {
        console.error('VerifyEmail - Error:', err);
        console.error('VerifyEmail - Error details:', {
          message: err.message,
          stack: err.stack,
        });
        setStatus('error');
        setMessage(err.message || 'Greška pri verifikaciji email-a. Proverite da je token validan i nije istekao.');
      }
    };

    verifyEmail();
  }, [searchParams, navigate]);

  return (
    <div className="container">
      <div className="card">
        <h2>Verifikacija Email-a</h2>
        {status === 'verifying' && (
          <div>
            <p>Verifikacija u toku...</p>
          </div>
        )}
        {status === 'success' && (
          <div className="success">
            <p>{message}</p>
            <p>Preusmeravanje na stranicu za prijavu...</p>
          </div>
        )}
        {status === 'error' && (
          <div className="error">
            <p>{message}</p>
            <button
              className="btn btn-primary"
              onClick={() => navigate('/login')}
              style={{ marginTop: '10px' }}
            >
              Idi na prijavu
            </button>
          </div>
        )}
      </div>
    </div>
  );
};

export default VerifyEmail;
