// xendit.js

// Helper to grab order details from localStorage
function getOrderDetails() {
    return JSON.parse(localStorage.getItem('orderDetails') || '{}');
}

// Build the cart array in the shape your backend expects
function buildCart() {
    const orderDetails = getOrderDetails();
    if (!orderDetails.items) return [];

    return orderDetails.items.map(item => ({
        id: item.menuCode || item.code || item.id, // Use your menu code field as the "product id"
        quantity: item.quantity || 1, // Keep track of how many the user wants
    }));
}

// Create an invoice using Xendit API
async function createInvoice() {
    const cart = buildCart();
    if (cart.length === 0) {
        alert("Your cart is empty!");
        return;
    }

    try {
        const response = await fetch("/api/xendit/invoices", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({ cart }),
        });

        const invoiceData = await response.json();

        if (invoiceData.invoice_url) {
            // Redirect the user to the Xendit payment page
            window.location.href = invoiceData.invoice_url;
        } else {
            const errorMessage = invoiceData.message || "Failed to create invoice.";
            throw new Error(errorMessage);
        }
    } catch (error) {
        console.error("Error creating Xendit invoice:", error);
        alert(`Error: ${error.message}`);
    }
}

// Attach event listener to the payment button
document.querySelector("#xendit-button").addEventListener("click", createInvoice);

// Helper to display messages
function resultMessage(message) {
    document.querySelector("#result-message").innerHTML = message;
}
