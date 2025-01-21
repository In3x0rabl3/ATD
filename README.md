# ATD

### Attemping to training myself and others on AI / LLM vulnerabilities. Be caution where you spin this application, I'm not at fault if you start hosting this on your Server lol... This is my first real attempt using Go to create a working application so please don't be brutal.

## **Overview**

This project showcases four key modules demonstrating common security vulnerabilities and adversarial attack techniques:

1. **Prompt Injection**
2. **Data Leakage**
3. **Data Poisoning**
4. **Supply Chain Attacks**

Each module highlights specific attack vectors and provides opportunities for learning about exploitation and mitigation.

The project runs on a Go-based backend with SSL enabled for secure communication.

---

## Modules

### **1. Prompt Injection**

- **Objective**: Exploit an AI model by crafting malicious prompts to manipulate responses and extract sensitive information.
- **Key Features**:
  - Enables interaction with the AI using crafted prompts.
  - Demonstrates how adversarial prompts can bypass intended restrictions.


---

### **2. Data Leakage**

- **Objective**: Upload a malicious image containing SQL commands to expose sensitive database details.
- **Key Features**:
  - Processes uploaded files without proper validation, leading to SQL injection.
  - Demonstrates the exposure of database structures and potentially sensitive information.

---

### **3. Data Poisoning**

- **Objective**: Corrupt the baseline dataset by injecting malicious data to alter the AI model's behavior.
- **Key Features**:
  - Accepts CSV uploads to simulate data poisoning.
  - Shows the impact of poisoned datasets on model predictions and integrity.
  - 
---

### **4. Supply Chain Attacks**

- **Objective**: Demonstrate the risks of compromised machine learning models.
- **Key Features**:
  - Allows uploading of a backdoored model file (`malicious_chatbot.pth`).
  - Simulates interaction with the compromised model to highlight adversarial behavior.

---

## **Setup Instructions**

### **Prerequisites**

1. **Programming Language**: Go 1.20+  
2. **Python Requirements** (for the **Supply Chain** module):
   - Python 3.8+
   - **Dependencies**:
     - `torch` (install via pip)
   - It's recommended to set up a virtual environment:
     ```bash
     python3 -m venv venv
     source venv/bin/activate
     pip install torch
     ```

3. **SSL Certificates**:
   - Generate `cert.pem` and `key.pem` files for SSL:
     ```bash
     openssl req -newkey rsa:2048 -nodes -keyout key.pem -x509 -days 365 -out cert.pem
     ```

4. **SQLite**:
   - Ensure SQLite is installed for the database.

---

### **Steps to Run**

1. **Clone the Repository**:
   ```bash
   git clone https://github.com/In3x0rabl3/atd.git
   cd atd
   go mod init atd
   go mod tidy
   go run main.go
   
2. **Enjoy**:   
```bash
https://127.0.0.1:8443
