async function fetchUpdatedProfile() {
  try {
    const response = await fetch("/fetch_user_metadata");
    if (!response.ok) throw new Error("Failed to fetch updated profile");

    const data = await response.json();
    updateDonationList(data.tags);
  } catch (error) {
    console.error("❌ Failed to refresh profile:", error);
  }
}
