# Contributing to JTF (Just Transfer File)

Thank you for your interest in contributing to JTF! We greatly appreciate your willingness to help improve the project. Here's how you can contribute:

## 1. Fork the Repository:
Start by forking the JTF repository on GitHub. This will create a copy of the project in your GitHub account.
@
## 2. Clone the Repository:
Next, clone the forked repository to your local machine using the following command:

```bash
git clone https://github.com/your-username/jtf.git
```

## 3. Create a Branch:
Create a new branch in your local repository to work on your changes. It's a good practice to give your branch a descriptive name related to the changes you'll be making.

```bash
git checkout -b my-feature-branch
```

## 4. How to Run JTF After Switching Branch:
After switching to your feature branch, follow these steps to run JTF with your changes:
### 4.1. Setup Environment Variables:
Copy the sample.env file to .env and fill in all the required fields. The .env file is used to store sensitive information and configuration settings. Here's what you need to fill in:
```bash # Get the key from https://ipinfo.io/ and paste it below
LOCATION_INFO_KEY=your_ipinfo_key

# Generate a secret key by running this command "go run ./pkg/random_key/main.go"
ENCRYPT_SECRET_KEY=your_secret_key

# Paste your encrypted private key here. To encrypt your private key, run the command "go run ./pkg/encrypt/main.go" and fill all required fields. Copy the encrypted key and paste it below.
ENCRYPTED_PRIVATE_KEY="your_encrypted_private_key"

# Go to your GitHub developer settings and create an application to get the client_id and secret_key. Set the redirect URL to http://localhost:3000/login/github/callback.
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_SECRET_KEY=your_github_secret_key
GITHUB_REDIRECT_URI=http://localhost:3000/login/github/callback

# Set the client_id and secret_key obtained from your Google developer settings.
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_SECRET_KEY=your_google_secret_key
GOOGLE_REDIRECT_URI=http://localhost:3000/login/google/callback
```
### 4.2. Run the Project:
Once you have successfully filled in the .env file, you can run the JTF project with the following command:
```bash
make run
```
JTF will start running, and you can access the application at the specified address (usually http://localhost:3000). Your changes and configurations from the feature branch will now be applied and reflected in the running application.

## 5. Make Changes:
Now, you can make the desired changes to the codebase. Feel free to refactor, add new features, fix bugs, or improve the documentation.

## 6. Test Your Changes:
Ensure that your changes are thoroughly tested to maintain the stability of the project.

## 7. Commit Your Changes:
Once you are satisfied with your changes, commit them with a clear and descriptive commit message.

```bash
git add .
git commit -m "Add new feature: XYZ
```

## 8. Push Changes:
Push your changes to your forked repository on GitHub.

```bash
git push origin my-feature-branch
```

## 9. Create a Pull Request:
Navigate to the original JTF repository on GitHub and create a Pull Request (PR) from your feature branch to main. Provide a detailed explanation of the changes you made in the PR description.

## 10. Review and Merge:
The maintainers of the JTF repository will review your PR. If any changes or improvements are required, they will let you know. Once approved, your changes will be merged into the main repository.

## 11. Thank You!
Congratulations on your successful contribution to JTF! Your efforts will benefit the entire community of users.

# Remember, open-source projects thrive on collaboration, and every contribution, no matter how big or small, is valuable. Happy coding!
