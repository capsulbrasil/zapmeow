## ZapMeow

ZapMeow is a versatile API that allows developers to interact with WhatsApp using the [whatsmeow](https://github.com/tulir/whatsmeow) library. This API was designed to facilitate communication and allow the use of multiple instances of WhatsApp.

### Features

- **Multi-Instance Support**: Seamlessly manage and interact with multiple WhatsApp instances concurrently.
- **Message Sending**: Send text, image, and audio messages to WhatsApp contacts and groups.
- **Phone Number Verification**: Check if phone numbers are registered on WhatsApp.
- **Contact Information**: Retrieve contact information, including names and emails.
- **Profile Information**: Obtain profile information for WhatsApp users.
- **QR Code Generation**: Generate QR codes to initiate WhatsApp login.
- **Instance Status**: Retrieve the connection status of a specific instance of WhatsApp.

### Getting Started

To get started with the ZapMeow API, follow these simple steps:

1. **Clone the Repository**: Clone this repository to your local machine using the following command:

   ```sh
   git clone git@github.com:capsulbrasil/zapmeow.git
   ```

2. **Configuration**: Set up your project configuration by copying the provided `.env.example` file and updating the environment variables.

   - Navigate to the project directory:

     ```sh
     cd zapmeow
     ```

   - Create a copy of the `.env.example` file as `.env`:

     ```sh
     cp .env.example .env
     ```

   - Open the `.env` file using your preferred text editor and update the necessary environment variables.

3. **Install Dependencies**: Install the project dependencies using the following command:

   ```sh
   go mod tidy
   ```

4. **Start the API**: Run the API server by executing the following command:

   ```sh
   go run main.go
   ```

5. **Access Swagger Documentation**: You can access the Swagger documentation by visiting the following URL in your web browser:

   ```
   http://localhost:8900/api/swagger/index.html
   ```

   The Swagger documentation provides detailed information about the available API endpoints, request parameters, and response formats.

Now, your ZapMeow API is up and running, ready for you to start interacting with WhatsApp instances programmatically.
