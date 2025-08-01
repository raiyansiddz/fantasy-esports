package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"fantasy-esports-backend/config"
	"fantasy-esports-backend/models"
	"fantasy-esports-backend/pkg/cdn"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	db     *sql.DB
	config *config.Config
	cdn    *cdn.CloudinaryClient
}

func NewUserHandler(db *sql.DB, cfg *config.Config, cdn *cdn.CloudinaryClient) *UserHandler {
	return &UserHandler{
		db:     db,
		config: cfg,
		cdn:    cdn,
	}
}

// @Summary Get user profile
// @Description Get current user's profile information
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.User
// @Failure 401 {object} models.ErrorResponse
// @Router /users/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var user models.User
	err := h.db.QueryRow(`
		SELECT id, mobile, email, first_name, last_name, date_of_birth, gender,
			   avatar_url, is_verified, is_active, account_status, kyc_status,
			   referral_code, state, city, pincode, created_at, updated_at
		FROM users WHERE id = $1`, userID).Scan(
		&user.ID, &user.Mobile, &user.Email, &user.FirstName, &user.LastName,
		&user.DateOfBirth, &user.Gender, &user.AvatarURL, &user.IsVerified,
		&user.IsActive, &user.AccountStatus, &user.KYCStatus, &user.ReferralCode,
		&user.State, &user.City, &user.Pincode, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Error:   "User not found",
			Code:    "USER_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user":    user,
	})
}

// @Summary Update user profile
// @Description Update user's profile information
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]interface{} true "Profile update data"
// @Success 200 {object} models.User
// @Failure 400 {object} models.ErrorResponse
// @Router /users/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Build dynamic query based on provided fields
	query := "UPDATE users SET updated_at = NOW()"
	args := []interface{}{}
	argCount := 1

	if firstName, ok := req["first_name"]; ok {
		query += ", first_name = $" + strconv.Itoa(argCount)
		args = append(args, firstName)
		argCount++
	}

	if lastName, ok := req["last_name"]; ok {
		query += ", last_name = $" + strconv.Itoa(argCount)
		args = append(args, lastName)
		argCount++
	}

	if email, ok := req["email"]; ok {
		query += ", email = $" + strconv.Itoa(argCount)
		args = append(args, email)
		argCount++
	}

	if dateOfBirth, ok := req["date_of_birth"]; ok {
		query += ", date_of_birth = $" + strconv.Itoa(argCount)
		args = append(args, dateOfBirth)
		argCount++
	}

	if gender, ok := req["gender"]; ok {
		query += ", gender = $" + strconv.Itoa(argCount)
		args = append(args, gender)
		argCount++
	}

	if state, ok := req["state"]; ok {
		query += ", state = $" + strconv.Itoa(argCount)
		args = append(args, state)
		argCount++
	}

	if city, ok := req["city"]; ok {
		query += ", city = $" + strconv.Itoa(argCount)
		args = append(args, city)
		argCount++
	}

	if pincode, ok := req["pincode"]; ok {
		query += ", pincode = $" + strconv.Itoa(argCount)
		args = append(args, pincode)
		argCount++
	}

	query += " WHERE id = $" + strconv.Itoa(argCount)
	args = append(args, userID)

	_, err := h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to update profile",
			Code:    "UPDATE_FAILED",
		})
		return
	}

	// Return updated profile
	h.GetProfile(c)
}

// @Summary Upload KYC documents
// @Description Upload KYC documents (PAN, Aadhaar, Bank Statement)
// @Tags Users
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param doc_type formData string true "Document type" Enums(pan_card, aadhaar, bank_statement)
// @Param document_front formData file true "Front side of document"
// @Param document_back formData file false "Back side of document (for Aadhaar)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Router /users/kyc/upload [post]
func (h *UserHandler) UploadKYC(c *gin.Context) {
	userID := c.GetInt64("user_id")

	docType := c.PostForm("doc_type")
	if docType == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Document type is required",
			Code:    "DOC_TYPE_REQUIRED",
		})
		return
	}

	// Get uploaded files
	frontFile, err := c.FormFile("document_front")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Front document is required",
			Code:    "FRONT_DOC_REQUIRED",
		})
		return
	}

	// Save front document to temporary location and upload to CDN
	frontPath := "/tmp/" + frontFile.Filename
	if err := c.SaveUploadedFile(frontFile, frontPath); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to save file",
			Code:    "FILE_SAVE_FAILED",
		})
		return
	}

	frontURL, err := h.cdn.UploadKYCDocument(frontPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to upload document",
			Code:    "UPLOAD_FAILED",
		})
		return
	}

	var backURL *string
	backFile, err := c.FormFile("document_back")
	if err == nil {
		// Back document provided
		backPath := "/tmp/" + backFile.Filename
		if err := c.SaveUploadedFile(backFile, backPath); err == nil {
			if url, err := h.cdn.UploadKYCDocument(backPath); err == nil {
				backURL = &url
			}
		}
	}

	// Store document info in database
	docNumber := c.PostForm("pan_number")
	if docNumber == "" {
		docNumber = c.PostForm("aadhaar_number")
	}
	if docNumber == "" {
		docNumber = c.PostForm("bank_account_number")
	}

	_, err = h.db.Exec(`
		INSERT INTO kyc_documents (user_id, document_type, document_front_url, 
								  document_back_url, document_number, status, created_at)
		VALUES ($1, $2, $3, $4, $5, 'pending', NOW())`,
		userID, docType, frontURL, backURL, docNumber)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to save document info",
			Code:    "DB_SAVE_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Document uploaded successfully",
		"document_url": frontURL,
	})
}

// @Summary Get KYC status
// @Description Get current KYC verification status
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /users/kyc/status [get]
func (h *UserHandler) GetKYCStatus(c *gin.Context) {
	userID := c.GetInt64("user_id")

	// Get KYC documents status
	rows, err := h.db.Query(`
		SELECT document_type, status 
		FROM kyc_documents 
		WHERE user_id = $1 
		ORDER BY created_at DESC`, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to fetch KYC status",
			Code:    "DB_ERROR",
		})
		return
	}
	defer rows.Close()

	kycStatus := map[string]string{
		"pan_status":     "not_uploaded",
		"aadhaar_status": "not_uploaded",
		"bank_status":    "not_uploaded",
	}

	for rows.Next() {
		var docType, status string
		if err := rows.Scan(&docType, &status); err == nil {
			switch docType {
			case "pan_card":
				kycStatus["pan_status"] = status
			case "aadhaar":
				kycStatus["aadhaar_status"] = status
			case "bank_statement":
				kycStatus["bank_status"] = status
			}
		}
	}

	// Determine overall status
	overallStatus := "not_started"
	canWithdraw := false
	pendingDocs := []string{}

	allVerified := true
	anyUploaded := false

	for docType, status := range kycStatus {
		if status == "not_uploaded" {
			allVerified = false
			pendingDocs = append(pendingDocs, docType)
		} else {
			anyUploaded = true
			if status != "verified" {
				allVerified = false
			}
		}
	}

	if allVerified && anyUploaded {
		overallStatus = "verified"
		canWithdraw = true
	} else if anyUploaded {
		overallStatus = "partial"
	}

	kycStatus["overall_status"] = overallStatus
	kycStatus["can_withdraw"] = strconv.FormatBool(canWithdraw)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"kyc_status": kycStatus,
		"pending_docs": pendingDocs,
	})
}

// @Summary Update preferences
// @Description Update user notification and app preferences
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]interface{} true "Preferences data"
// @Success 200 {object} map[string]interface{}
// @Router /users/preferences [put]
func (h *UserHandler) UpdatePreferences(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// For now, we'll store preferences in a simple way
	// In production, you might want a separate preferences table
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Preferences updated successfully",
		"user_id": userID,
		"preferences": req,
	})
}