package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/utils"
)

// An end point to create new users
func Create(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	user_collection := "users"

	var user User

	err := utils.ParseJsonFromRequest(request, &user)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, response)
		return
	}
	if !utils.IsValidEmail(user.Email) {
		utils.GetError(errors.New("email address is not valid"), http.StatusBadRequest, response)
		return
	}

	// confirm if user_email exists
	result, _ := utils.GetMongoDbDoc(user_collection, bson.M{"email": user.Email})
	if result != nil {
		fmt.Printf("users with email %s exists!", user.Email)
		utils.GetError(errors.New("operation failed"), http.StatusBadRequest, response)
		return
	}

	user.CreatedAt = time.Now()

	detail, _ := utils.StructToMap(user)

	res, err := utils.CreateMongoDbDoc(user_collection, detail)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
		return
	}

	utils.GetSuccess("user created", res, response)
}

// helper functions perform CRUD operations on user
func FindUserByID(response http.ResponseWriter, request *http.Request) {
	// Find a user by user ID
	response.Header().Set("content-type", "application/json")

	collectionName := "users"
	userID := mux.Vars(request)["id"]
	objID, err := primitive.ObjectIDFromHex(userID)

	if err != nil {
		utils.GetError(err, http.StatusBadRequest, response)
		return
	}

	res, err := utils.GetMongoDbDoc(collectionName, bson.M{"_id": objID})
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
		return
	}
	utils.GetSuccess("User retrieved successfully", res, response)

}

func UpdateUser(response http.ResponseWriter, request *http.Request) {
	// Update a user of a given ID. Only certain fields, detailed in the
	// UserUpdate struct can be directly updated by a user without additional
	// functionality or permissions
	response.Header().Set("content-type", "application/json")

	collectionName := "users"
	userID := mux.Vars(request)["id"]
	objID, err := primitive.ObjectIDFromHex(userID)
	// Validate the user ID provided
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, response)
		return
	}

	res, err := utils.GetMongoDbDoc(collectionName, bson.M{"_id": objID})
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
		return
	}
	if res != nil {
		// 2. Get user fields to be updated from request body
		var body UserUpdate
		err := json.NewDecoder(request.Body).Decode(&body)
		if err != nil {
			utils.GetError(err, http.StatusBadRequest, response)
			return
		}

		// Convert body struct to interface
		var userInterface map[string]interface{}
		bytes, err := json.Marshal(body)
		if err != nil {
			utils.GetError(err, http.StatusInternalServerError, response)
		}
		json.Unmarshal(bytes, &userInterface)

		// 3. Update user
		updateRes, err := utils.UpdateOneMongoDbDoc(collectionName, userID, userInterface)
		if err != nil {
			utils.GetError(err, http.StatusInternalServerError, response)
			return
		}
		utils.GetSuccess("User update successful", updateRes, response)
	}

}
