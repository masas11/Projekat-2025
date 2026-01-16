// Simple encryption/decryption for localStorage data
// Uses AES-like encryption with Base64 encoding
// Note: This is client-side encryption for basic protection, not cryptographically secure
// For production, use proper encryption libraries or server-side token management

const ENCRYPTION_KEY = 'default-encryption-key-change-in-production'; // In production, generate dynamically or use proper key management

// Simple XOR encryption (basic obfuscation for demonstration)
// For real security, use Web Crypto API or proper encryption library
function encrypt(text) {
  if (!text) return '';
  
  try {
    // Simple Base64 encoding with key obfuscation for demonstration
    // In production, use Web Crypto API: crypto.subtle.encrypt()
    const encoded = btoa(encodeURIComponent(text));
    let encrypted = '';
    for (let i = 0; i < encoded.length; i++) {
      const charCode = encoded.charCodeAt(i) ^ ENCRYPTION_KEY.charCodeAt(i % ENCRYPTION_KEY.length);
      encrypted += String.fromCharCode(charCode);
    }
    return btoa(encrypted);
  } catch (error) {
    console.error('Encryption error:', error);
    return text; // Fallback to plain text if encryption fails
  }
}

function decrypt(encryptedText) {
  if (!encryptedText) return '';
  
  try {
    // Decrypt Base64 encoded text
    let decrypted = atob(encryptedText);
    let decoded = '';
    for (let i = 0; i < decrypted.length; i++) {
      const charCode = decrypted.charCodeAt(i) ^ ENCRYPTION_KEY.charCodeAt(i % ENCRYPTION_KEY.length);
      decoded += String.fromCharCode(charCode);
    }
    return decodeURIComponent(atob(decoded));
  } catch (error) {
    console.error('Decryption error:', error);
    return encryptedText; // Fallback to encrypted text if decryption fails
  }
}

// Store encrypted data in localStorage
export function setEncryptedItem(key, value) {
  try {
    const jsonValue = JSON.stringify(value);
    const encrypted = encrypt(jsonValue);
    localStorage.setItem(key, encrypted);
    // Also store a checksum for integrity verification
    const checksum = btoa(JSON.stringify(value).split('').reduce((acc, char) => acc + char.charCodeAt(0), 0));
    localStorage.setItem(key + '_checksum', checksum);
  } catch (error) {
    console.error('Error storing encrypted item:', error);
    // Fallback to plain storage if encryption fails
    localStorage.setItem(key, JSON.stringify(value));
  }
}

// Retrieve and decrypt data from localStorage with integrity check
export function getEncryptedItem(key) {
  try {
    const encrypted = localStorage.getItem(key);
    if (!encrypted) return null;
    
    const decrypted = decrypt(encrypted);
    if (!decrypted) return null;
    
    const parsed = JSON.parse(decrypted);
    
    // Verify integrity
    const storedChecksum = localStorage.getItem(key + '_checksum');
    if (storedChecksum) {
      const currentChecksum = btoa(JSON.stringify(parsed).split('').reduce((acc, char) => acc + char.charCodeAt(0), 0));
      if (storedChecksum !== currentChecksum) {
        console.warn('Integrity check failed for', key);
        // Data may have been tampered with
        localStorage.removeItem(key);
        localStorage.removeItem(key + '_checksum');
        return null;
      }
    }
    
    return parsed;
  } catch (error) {
    console.error('Error retrieving encrypted item:', error);
    // Try to get plain item as fallback
    const plain = localStorage.getItem(key);
    if (plain) {
      try {
        return JSON.parse(plain);
      } catch {
        return plain;
      }
    }
    return null;
  }
}

// Remove encrypted item
export function removeEncryptedItem(key) {
  localStorage.removeItem(key);
  localStorage.removeItem(key + '_checksum');
}
