import pandas as pd
import numpy as np
import pickle
from sklearn.model_selection import train_test_split
from sklearn.linear_model import LinearRegression
from sklearn.metrics import mean_absolute_error

# load dataset
df = pd.read_csv("porsche_data.csv")

# select features and target variable
X = df[["speed", "engine_temp"]]
y = df[["fuel_efficiency"]]

# split data innto training and test sets
X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2, random_state=42)


# train linear regression model
model = LinearRegression()
model.fit(X_train, y_train)

# evaluate the model
predictions = model.predict(X_test)
mse = mean_squared_error(y_test, predictions)
print(f"Model Mean Squared Error: {mse}")

# Save model to a file
with open("fuel_efficiency_model.pkl", "wb") as f:
    pickle.dump(model, f)