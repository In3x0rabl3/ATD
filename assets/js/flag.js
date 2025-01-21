document.addEventListener("DOMContentLoaded", () => {
    const flagInput = document.getElementById("flag-input");
    const flagSubmitBtn = document.getElementById("flag-submit-btn");
    const flagResponse = document.getElementById("flag-response");
    const moduleSelect = document.getElementById("module-select");

    flagSubmitBtn.addEventListener("click", () => {
        const flag = flagInput.value.trim();
        const module = moduleSelect.value;

        // Clear previous response styles
        flagResponse.classList.remove("success", "error");

        if (!flag) {
            flagResponse.textContent = "Please enter a flag!";
            flagResponse.classList.add("error"); // Add error class
            return;
        }

        if (!module) {
            flagResponse.textContent = "Please select a module!";
            flagResponse.classList.add("error"); // Add error class
            return;
        }

        // Submit the flag and module to the server
        fetch("/submit-flag", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({ module, flag }), // Include the module and flag
        })
            .then((response) => {
                if (!response.ok) {
                    throw new Error("Server returned an error");
                }
                return response.json();
            })
            .then((data) => {
                if (data.message.includes("Congratulations")) {
                    flagResponse.textContent = data.message;
                    flagResponse.classList.add("success"); // Add success class
                } else {
                    flagResponse.textContent = data.message;
                    flagResponse.classList.add("error"); // Add error class
                }
            })
            .catch((error) => {
                flagResponse.textContent = "Error submitting flag!";
                flagResponse.classList.add("error"); // Add error class
                console.error("Error:", error);
            });
    });
});
