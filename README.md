---

### **TheVinBoxProject: An Email Summarizer** 🚗💨

This is a Go application for the crew who have no time for long email chains. It connects to your Gmail account, grabs the latest messages, and gives you the rundown, the essentials, in a way that gets right to the point, with a little attitude to match. It’s not just about the message; it's about the mission.

---

### **Key Features** ✨

* **Secure Gmail Integration**: Authenticates with the Gmail API to securely access and read your inbox.
* **Email Processing**: Scans your emails for subject and body content, focusing on the core message.
* **Plaintext Extraction**: Strips HTML tags and other clutter to get to the truth of the message.
* **Dominic Toretto Persona**: Summarizes the email content in the straight-talking, family-first style of Dom Toretto.

---

### **Prerequisites** 📋

Before you get this ride on the road, you'll need a few things:

* **Go**: Version 1.18 or higher.
* **Gmail API Credentials**: A `credentials.json` file from the Google Cloud Console with the Gmail API enabled. You'll need to set up OAuth 2.0 to handle authentication.

---

### **Installation** 🔧

1.  **Clone the Repository**:
    ```sh
    git clone https://github.com/tysenh1/TheVinBoxProjectGo
    cd TheVinBoxProjectGo
    ```
2.  **Get Your Gmail API Credentials**:
    Follow the steps in the [Google API documentation](https://developers.google.com/gmail/api/quickstart/go) to create a project, enable the Gmail API, and download your `credentials.json` file. Place this file in the same directory as your `main.go` file.
3.  **Install Dependencies**:
    ```sh
    go mod tidy
    ```

---

### **Usage** 🗣️

To see the application in action, simply run the main file. The first time you run it, you'll be prompted to authenticate with your Google account in your browser.

```sh
go run main.go
