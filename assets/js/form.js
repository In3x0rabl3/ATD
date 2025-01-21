import { fetchScore } from "./score.js";
import { clearInputs } from "./utils.js";
import { updateScoreBox } from "./score.js";

document.addEventListener("DOMContentLoaded", () => {
    const form = document.querySelector("form");
    const responseContent = document.getElementById("response-content");
    const promptElement = document.getElementById("prompt");
    const fileInput = document.getElementById("file");
    const modelInput = document.getElementById("modell");

    let isSubmitting = false; // Prevent double submits

    function handleFormSubmit(event) {
        event.preventDefault();

        // Check and lock submission
        if (isSubmitting) return;
        isSubmitting = true;

        const formData = new FormData(form);
        const selectedModel = formData.get("model");

        responseContent.innerText = "";

        if (!selectedModel) {
            responseContent.innerText = "Please select a model.";
            isSubmitting = false;
            return;
        }

        const file = fileInput?.files[0];
        const model = modelInput?.files[0];
        const prompt = promptElement?.value.trim();

        if (selectedModel === "Data Leakage") {
            if (file && file.type.startsWith("image/*")) {
                formData.append("type", "image");
                responseContent.innerText = "Uploading image...";
                isSubmitting = false;
                return;
            }
        } else if (selectedModel === "Data Poisoning") {
            if (file) {
                formData.append("type", "file");
                responseContent.innerText = "Uploading dataset...";
            } else if (prompt) {
                formData.append("type", "prompt");
                responseContent.innerText = "Processing prompt...";
            } else {
                responseContent.innerText = "Please upload a file or enter a prompt.";
                isSubmitting = false;
                return;
            }
        } else if (["Prompt Injection"].includes(selectedModel)) {
            if (prompt) {
                formData.append("type", "prompt");
                responseContent.innerText = "Processing prompt...";
            } else {
                responseContent.innerText = "Please enter a prompt.";
                isSubmitting = false;
                return;
            }
        } else if (selectedModel === "Supply Chain") {
            if (model) {
                formData.append("type", "modell");
                responseContent.innerText = "Uploading model...";
            } else if (prompt) {
                // Interact with LLM
                fetch("/supply-chain/chat", {
                    method: "POST",
                    headers: {
                        "Content-Type": "application/json",
                    },
                    body: JSON.stringify({ prompt }),
                })
                    .then((response) => {
                        if (!response.ok) {
                            throw new Error(`HTTP error! Status: ${response.status}`);
                        }
                        return response.json();
                    })
                    .then((data) => {
                        responseContent.innerText =
                            data.response || data.message || "LLM response received.";
                    })
                    .catch((error) => {
                        responseContent.innerText =
                            "Error interacting with LLM. Please try again.";
                        console.error("Error interacting with LLM:", error);
                    })
                    .finally(() => {
                        isSubmitting = false;
                    });
                return;
            } else {
                responseContent.innerText =
                    "Please upload a model or enter a prompt.";
                isSubmitting = false;
                return;
            }
        } else {
            responseContent.innerText = "Unsupported model selected.";
            isSubmitting = false;
            return;
        }

        fetch("/submit", { method: "POST", body: formData })
            .then((response) => {
                if (!response.ok) {
                    throw new Error(`HTTP error! Status: ${response.status}`);
                }
                return response.json();
            })
            .then((data) => {
                responseContent.innerText =
                    data.response || data.message || "Submission successful!";

                // Update score after successful submission
                fetchScore();
            })
            .catch((error) => {
                responseContent.innerText =
                    "Error submitting form. Please try again.";
                console.error("Error submitting form:", error);
            })
            .finally(() => {
                isSubmitting = false; // Reset submission state
                clearInputs();
            });
    }

    function handleKeyDown(event) {
        if (event.key === "Enter" && event.target === promptElement) {
            event.preventDefault(); // Prevent default Enter behavior in textarea
            form.requestSubmit(); // Trigger form submit programmatically
        }
    }

    fetchScore();

    // Remove any existing listener to avoid duplicates
    form.removeEventListener("submit", handleFormSubmit);

    // Attach event listeners
    form.addEventListener("submit", handleFormSubmit);
    if (promptElement) {
        promptElement.addEventListener("keydown", handleKeyDown);
    }
});
