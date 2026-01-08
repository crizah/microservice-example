// page where user enters token and password
// sends token and password to badkend for verification 
// backend verifies 
// after successful change, redirect to login page



import React, { useState } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import axios from "axios";
import "./ResetPassword.css";

export function ResetPassword() {
  const navigate = useNavigate();
  const location = useLocation();
  const email = location.state?.email || "";
  
  const [formData, setFormData] = useState({
    token: "",
    newPassword: "",
    confirmPassword: ""
  });
  const [message, setMessage] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();

   
    if (!formData.token) {
      setMessage("Please enter the reset code");
      setIsSuccess(false);
      return;
    }

    if (!formData.newPassword) {
      setMessage("Please enter new password");
      setIsSuccess(false);
      return;
    }


    if (formData.newPassword !== formData.confirmPassword) {
      setMessage("Passwords don't match");
      setIsSuccess(false);
      return;
    }

    setIsLoading(true);
    setMessage("");

    try {
      const res = await axios.post("http://localhost:8080/reset-password", {
        email: email,
        token: formData.token,
        newPassword: formData.newPassword
      });

      setIsSuccess(true);
      setMessage("Password reset successfully! Redirecting to login...");

      // Redirect to login page
      setTimeout(() => {
        navigate("/login");
      }, 2000);

    } catch (error) {
      setIsSuccess(false);

      if (error.response) {
        setMessage(error.response.data || "Invalid or expired code");
      } else if (error.request) {
        setMessage("Cannot connect to server. Please try again later.");
      } else {
        setMessage("An error occurred. Please try again.");
      }

      console.error("Error:", error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleResendCode = async () => {
    setIsLoading(true);
    setMessage("");

    try {
      await axios.post("http://localhost:8080/request-password-reset", {
        email: email
      });

      setIsSuccess(true);
      setMessage("Reset code resent to your email");
    } catch (error) {
      setMessage("Failed to resend code");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="reset-password-container">
      <div className="reset-password-card">
        <div className="key-icon">🔑</div>
        
        <h2>Reset Password</h2>
        
        <p className="reset-instruction">
          Enter the 6-digit code sent to <strong>{email}</strong> and your new password
        </p>

        <form onSubmit={handleSubmit}>
          <div className="input-group">
            <label>Reset Code</label>
            <input
              type="text"
              placeholder="Enter 6-digit code"
              value={formData.token}
              onChange={(e) => setFormData({ ...formData, token: e.target.value })}
              maxLength="6"
              disabled={isLoading}
              required
            />
          </div>

          <div className="input-group">
            <label>New Password</label>
            <input
              type="password"
              placeholder="Enter new password (min 8 characters)"
              value={formData.newPassword}
              onChange={(e) => setFormData({ ...formData, newPassword: e.target.value })}
              disabled={isLoading}
              required
            />
          </div>

          <div className="input-group">
            <label>Confirm Password</label>
            <input
              type="password"
              placeholder="Confirm new password"
              value={formData.confirmPassword}
              onChange={(e) => setFormData({ ...formData, confirmPassword: e.target.value })}
              disabled={isLoading}
              required
            />
          </div>

          <button type="submit" disabled={isLoading} className="submit-button">
            {isLoading ? "Resetting Password..." : "Reset Password"}
          </button>

          {message && (
            <p className={`message ${isSuccess ? "success" : "error"}`}>
              {message}
            </p>
          )}
        </form>

        <div className="resend-section">
          <p>Didn't receive the code?</p>
          <button 
            onClick={handleResendCode} 
            disabled={isLoading}
            className="resend-button"
          >
            Resend Code
          </button>
        </div>

        <div className="back-to-login">
          <p>
            Remember your password?{" "}
            <a href="/login" onClick={(e) => {
              e.preventDefault();
              navigate("/login");
            }}>
              Back to Login
            </a>
          </p>
        </div>
      </div>
    </div>
  );
}