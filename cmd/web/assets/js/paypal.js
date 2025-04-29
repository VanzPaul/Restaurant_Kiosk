// paypal.js

// Helper to grab order details from localStorage
function getOrderDetails() {
    return JSON.parse(localStorage.getItem('orderDetails') || '{}');
}

// Build the cart array in the shape your backend expects
function buildCart() {
    const orderDetails = getOrderDetails();
    if (!orderDetails.items) return [];

    return orderDetails.items.map(item => ({
        // Use your menu code field as the "product id"
        id: item.menuCode || item.code || item.id,
        // Keep track of how many the user wants
        quantity: item.quantity || 1,
    }));
}

const paypalButtons = window.paypal.Buttons({
    style: {
         shape: "rect",
         layout: "vertical",
         color: "gold",
         label: "paypal",
     },

    // Create Order – now sends your real cart
    async createOrder() {
         const cart = buildCart();
         if (cart.length === 0) {
             throw new Error("Your cart is empty!");
         }

         const response = await fetch("/api/orders", {
             method: "POST",
             headers: {
                 "Content-Type": "application/json",
             },
             body: JSON.stringify({ cart }),
         });

         const orderData = await response.json();

         if (orderData.id) {
             return orderData.id;
         }

         const errorDetail = orderData?.details?.[0];
         const errorMessage = errorDetail
             ? `${errorDetail.issue} ${errorDetail.description} (${orderData.debug_id})`
             : JSON.stringify(orderData);

         throw new Error(errorMessage);
     },

    // Capture funds on approval (unchanged)
    async onApprove(data, actions) {
         try {
             const response = await fetch(
                 `/api/orders/${data.orderID}/capture`,
                 {
                     method: "POST",
                     headers: {
                         "Content-Type": "application/json",
                     },
                 }
             );

             const orderData = await response.json();
             const errorDetail = orderData?.details?.[0];

             if (errorDetail?.issue === "INSTRUMENT_DECLINED") {
                 return actions.restart();
             } else if (errorDetail) {
                 throw new Error(
                     `${errorDetail.description} (${orderData.debug_id})`
                 );
             } else if (!orderData.purchase_units) {
                 throw new Error(JSON.stringify(orderData));
             } else {
                 const transaction =
                     orderData.purchase_units[0].payments.captures?.[0] ||
                     orderData.purchase_units[0].payments.authorizations?.[0];
                 resultMessage(
                     `Transaction ${transaction.status}: ${transaction.id}<br>
                     <br>See console for all available details`
                 );
                 console.log("Capture result", orderData, JSON.stringify(orderData, null, 2));
             }
         } catch (error) {
             console.error(error);
             resultMessage(
                 `Sorry, your transaction could not be processed...<br><br>${error}`
             );
         }
     },
});

paypalButtons.render("#paypal-button-container");

// Helper to display messages
function resultMessage(message) {
    document.querySelector("#result-message").innerHTML = message;
}
