import React, { useState } from "react";
// import { useState } from "react";

import { useNavigate } from "react-router-dom";
import axios from "axios";
import "./LogIn.css"


// need a password reset request handler
// when clicked, goes to a page to enter their email
// send that email to backend

export function Login() {
  const navigate = useNavigate();
  const [formData, setFormData] = useState({
    identifier: "", // Can be email or username
    password: ""
  });
  const [message, setMessage] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsLoading(true);
    setMessage("");

    const isEmail = formData.identifier.includes("@");
    
    try {
      const res = await axios.post(
        "http://localhost:8080/login",
        isEmail
          ? { email: formData.identifier, password: formData.password }
          : { username: formData.identifier, password: formData.password },
        { withCredentials: true } // Sends cookies
      );

      setMessage("Login successful!");
      
      // Redirect to dashboard/home
      setTimeout(() => {
        navigate("/dashboard");
      }, 1000);
      
    } catch (error) {
      if (error.response) {
        setMessage(error.response.data || "Login failed");
      } else {
        setMessage("Cannot connect to server");
      }
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="login-container">
      <h2>Login</h2>
      <form onSubmit={handleSubmit}>
        <input
          type="text"
          placeholder="Username or Email"
          value={formData.identifier}
          onChange={(e) =>
            setFormData({ ...formData, identifier: e.target.value })
          }
          disabled={isLoading}
        />
        <input
          type="password"
          placeholder="Password"
          value={formData.password}
          onChange={(e) =>
            setFormData({ ...formData, password: e.target.value })
          }
          disabled={isLoading}
        />
        <button type="submit" disabled={isLoading}>
          {isLoading ? "Logging in..." : "Login"}
        </button>
        {message && <p>{message}</p>}
      </form>
    </div>
  );
}