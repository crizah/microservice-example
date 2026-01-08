import { useState } from "react";
import { useNavigate } from "react-router-dom";
import axios from "axios";
import "./SignUp.css";


export function SignUp() {
    
//   const fetchUserData = async () => { 
//     navigate = useNavigate();// used for every page that is accessed after login
//   try {
//     const res = await axios.get("http://localhost:8080/profile", {
//       withCredentials: true // Sends session cookie
//     });
//     console.log(res.data);
//   } catch (error) {
//     if (error.response?.status === 401) {
//       // Redirect to login
//       navigate("/login");
//     }
//   }
// };


  const [formData, setFormData] = useState({
    username: "",
    email: "",
    password: ""
  });
  const [message, setMessage] = useState("");
  const [isSuccess, setIsSuccess] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const navigate = useNavigate();

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: value
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    const { username, email, password } = formData;
    
   
    if (!username || !email || !password) {
      setIsSuccess(false);
      setMessage("All fields are required.");
      return;
    }

    // Send to backend
    setIsLoading(true);
    setMessage("");

    let config = {
      headers: {
        "Content-Type": "application/json"
        
      }
    }
    
    try {
      const res = await axios.post("http://localhost:8080/signup", {
        "username": formData.username,
        "email": formData.email,
        "password": formData.password
      }, config);
      
      console.log("Server response:", res.data);
      
      setIsSuccess(true);
      setMessage("Signup successful! Please check your email for verification.");
      
      // Clear form
      setFormData({
        username: "",
        email: "",
        password: ""
      });

      
      // Redirect after 1.5 seconds
      setTimeout(() => {
        navigate("/resend", { 
          state: { 
            email: formData.email,
            username: formData.username ,
            link : res.data.link
          } 
        });
      }, 1500);
      

      
    } catch (error) {
      setIsSuccess(false);
      
      if (error.response) {
        // Server responded with error
        setMessage(error.response.data || "Signup failed. Please try again.");
      } else if (error.request) {
        // No response from server
        setMessage("Cannot connect to server. Please try again later.");
      } else {
        setMessage("An error occurred. Please try again.");
      }
      
      console.error("Error:", error);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="signup-container">
      <h2>Sign Up</h2>
      <form onSubmit={handleSubmit}>
        <div className="input-group">
          <label>Username</label>
          <input
            type="text"
            name="username"
            value={formData.username}
            onChange={handleChange}
            disabled={isLoading}
          />
        </div>
        
        <div className="input-group">
          <label>Email</label>
          <input
            type="email"
            name="email"
            value={formData.email}
            onChange={handleChange}
            disabled={isLoading}
          />
        </div>
        
        <div className="input-group">
          <label>Password</label>
          <input
            type="password"
            name="password"
            value={formData.password}
            onChange={handleChange}
            disabled={isLoading}
          />
        </div>
        
        <button type="submit" disabled={isLoading}>
          {isLoading ? "Creating Account..." : "Create Account"}
        </button>
        
        {message && (
          <p className={isSuccess ? "success" : "error"}>
            {message}
          </p>
        )}
      </form>
      
      <p className="login-link">
        Already have an account? <a href="/login">Log in</a>
      </p>
    </div>
  );
}