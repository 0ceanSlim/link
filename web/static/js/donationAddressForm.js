document
  .getElementById("donation-form")
  .addEventListener("submit", async (event) => {
    event.preventDefault();

    const formData = new FormData(event.target);
    let newDonationTags = [];

    for (let i = 0; formData.has(`ticker_${i}`); i++) {
      const ticker = formData.get(`ticker_${i}`).trim();
      const network = formData.get(`network_${i}`).trim();
      const address = formData.get(`address_${i}`).trim();

      if (ticker && address) {
        if (network) {
          newDonationTags.push(["w", ticker, address, network]);
        } else {
          newDonationTags.push(["w", ticker, address]);
        }
      }
    }

    if (newDonationTags.length === 0) {
      alert("Please enter at least one donation address.");
      return;
    }

    if (!window.nostr) {
      alert("Nostr extension not found! Install a Nostr extension.");
      return;
    }

    try {
      const profileEvent = await fetchCurrentKind0FromCache();
      if (!profileEvent || !profileEvent.tags) {
        alert("Failed to fetch existing donation addresses.");
        return;
      }

      let updatedTags = [...profileEvent.tags];

      let existingDonationTags = updatedTags.filter((tag) => tag[0] === "w");

      newDonationTags.forEach((newTag) => {
        if (
          !existingDonationTags.some(
            (existingTag) =>
              JSON.stringify(existingTag) === JSON.stringify(newTag)
          )
        ) {
          existingDonationTags.push(newTag);
        }
      });

      updatedTags = updatedTags.filter((tag) => tag[0] !== "w"); // Remove old "w" tags
      updatedTags.push(...existingDonationTags); // Append updated tags

      const updatedEvent = {
        kind: 0,
        created_at: Math.floor(Date.now() / 1000),
        tags: updatedTags,
        content: profileEvent.content || "",
      };

      const signedEvent = await window.nostr.signEvent(updatedEvent);

      const response = await fetch("/save_donation_addresses", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(signedEvent),
      });

      if (!response.ok) {
        throw new Error(`Failed to save addresses: ${await response.text()}`);
      }

      alert("Donation address added successfully!");
      await fetchUpdatedProfile();

      setTimeout(() => {
        window.location.reload();
      }, 500);
    } catch (error) {
      console.error("❌ Error updating donation addresses:", error);
      alert(`Error: ${error.message}`);
    }
  });


let count = 1;

function addField() {
  const container = document.getElementById("donation-fields");
  const div = document.createElement("div");
  div.className =
    "flex flex-col gap-2 p-4 rounded-lg shadow donation-group bg-bgSecondary";
  div.innerHTML = `
      <input type="text" name="ticker_${count}" placeholder="Asset Ticker (e.g., BTC)" class="p-2 border rounded border-bgInverted focus:ring focus:ring-accent" required>
      <input type="text" name="network_${count}" placeholder="Network (e.g., Bitcoin)" class="p-2 border rounded border-bgInverted focus:ring focus:ring-accent" required>
      <input type="text" name="address_${count}" placeholder="Receiving Address" class="p-2 border rounded border-bgInverted focus:ring focus:ring-accent" required>
      <button type="button" onclick="removeField(this)" class="px-4 py-1 mt-2 text-sm font-semibold text-white bg-red-500 rounded-lg shadow hover:bg-red-600">Remove</button>
    `;
  container.appendChild(div);
  count++;
}

function removeField(button) {
  button.parentElement.remove();
}

async function fetchCurrentKind0FromCache() {
  try {
    const response = await fetch("/fetch_current_kind0");
    if (!response.ok) {
      throw new Error(`Failed to fetch Kind 0 data: ${await response.text()}`);
    }
    const data = await response.json();
    return JSON.parse(data.rawUserMetadataContent); // Convert string back to JSON
  } catch (error) {
    console.error("❌ Error fetching Kind 0 event from cache:", error);
    return null;
  }
}
