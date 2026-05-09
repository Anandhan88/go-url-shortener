package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminLoginPage renders the login page
func AdminLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin_login.html", gin.H{})
}

// AdminLogin handles the login form submission
func AdminLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	// Hardcoded credentials for simplicity
	if username == "admin" && password == "admin123" {
		// Set a simple cookie (in a real app, use sessions or JWT)
		c.SetCookie("admin_token", "authenticated", 3600*24, "/", "", false, true)
		c.Redirect(http.StatusFound, "/admin/dashboard")
		return
	}

	c.HTML(http.StatusUnauthorized, "admin_login.html", gin.H{
		"Error": "Invalid username or password",
	})
}

// AdminLogout clears the admin session
func AdminLogout(c *gin.Context) {
	c.SetCookie("admin_token", "", -1, "/", "", false, true)
	c.Redirect(http.StatusFound, "/admin/login")
}

// AdminAuthMiddleware protects routes requiring admin access
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("admin_token")
		if err != nil || token != "authenticated" {
			c.Redirect(http.StatusFound, "/admin/login")
			c.Abort()
			return
		}
		c.Next()
	}
}
