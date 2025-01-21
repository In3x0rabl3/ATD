import { clearInputs } from "./utils.js";

document.addEventListener("DOMContentLoaded", () => {
    const modelSelect = document.getElementById("model");
    const promptElement = document.getElementById("prompt");
    const responseContent = document.getElementById("response-content");
    const resetControls = document.getElementById("reset-controls");
    const Downloadmod = document.getElementById("download-model");
    const uploadSections = {
        image: document.getElementById("upload-image"),
        file: document.getElementById("upload-file"),
        model: document.getElementById("upload-model"),
    };

    const modelDescriptions = {
        "Prompt Injection": "Attempt to manipulate the AI using crafted prompt inputs to obtain the flag.",
        "Data Leakage": "Explore weaknesses that expose sensitive data by uploading images to reveal the hidden flag within the database. Can you query the database through an image?",
        "Data Poisoning": "Attempt to poison the baseline dataset to obtain the flag. Ask questions about the current dataset, develop your own malicious dataset. The flag will is injected into the dataset once the integrity level falls below 0.25, once you've succeeded convince the AI to hand over the flag. ",
        "Supply Chain": "Download the python script, generate a backdoored model, upload and interact to obtain the flag. Start the VM if not already started.",
    };

    // Function to clear response content
    function clearResponse() {
        if (responseContent) {
            responseContent.textContent = "No response yet.";
        }
    }

    // Function to update the UI based on the selected model
    function updateUI(selectedModel) {
        console.log("Updating UI for model:", selectedModel);

        // Clear inputs and response
        clearInputs();
        clearResponse();

        // Update instructions text
        document.getElementById("instructions").textContent =
            modelDescriptions[selectedModel] || "Select a module to view specific instructions.";

        // Toggle reset controls visibility
        resetControls.style.display = selectedModel === "Data Poisoning" ? "flex" : "none";

        // Toggle prompt textarea visibility
        promptElement.style.display = selectedModel === "Data Leakage" ? "none" : "block";


        Downloadmod.style.display = selectedModel === "Supply Chain" ? "flex" : "none";

        // Toggle upload sections visibility
        Object.values(uploadSections).forEach((section) => (section.style.display = "none"));
        if (selectedModel === "Data Leakage") {
            uploadSections.image.style.display = "block"; // Show image upload
        } else if (selectedModel === "Data Poisoning") {
            uploadSections.file.style.display = "block"
        } else if (selectedModel === "Supply Chain") {
            uploadSections.model.style.display = "block"
        }
    }

    // Event listener for model selection
    modelSelect.addEventListener("change", () => {
        const selectedModel = modelSelect.value;
        updateUI(selectedModel);
    });

    // Initialize UI on page load
    updateUI(modelSelect.value);
});
