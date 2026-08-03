<script lang="ts">
	import Modal from "../Modal.svelte";
	import Setting from "../settings/Setting.svelte";
	import SettingsList from "../settings/SettingsList.svelte";
	import { notify } from "../util/notify";
	import { req } from "../util/api";
	import type {
		CinemaDetailsUpdateRequest,
		CinemaScreen,
		GeocodeResult,
		WatchSource,
	} from "@/types";

	interface Props {
		source: WatchSource;
		onClose: () => void;
	}

	let { source, onClose }: Props = $props();

	let city = $state(source.cinema?.city ?? "");
	let address = $state(source.cinema?.address ?? "");
	let lat = $state(source.cinema?.lat);
	let lon = $state(source.cinema?.lon);
	let note = $state(source.cinema?.note ?? "");
	let ratingOverall = $state(source.cinema?.ratingOverall);
	let ratingSnacks = $state(source.cinema?.ratingSnacks);
	let ratingTech = $state(source.cinema?.ratingTech);
	let ratingComfort = $state(source.cinema?.ratingComfort);
	let screens = $state<CinemaScreen[]>(source.cinema?.screens ?? []);
	let newScreenName = $state("");
	let showExtraRatings = $state(false);
	let geocodeResults = $state<GeocodeResult[]>([]);
	let geocodeRunning = $state(false);
	let error = $state("");
	let submitDisabled = $state(false);

	async function geocode() {
		const query = [source.name, city].filter(Boolean).join(", ");
		geocodeRunning = true;
		try {
			geocodeResults = await req.get<GeocodeResult[]>(
				`/source/geocode?q=${encodeURIComponent(query)}`,
			);
			if (geocodeResults.length === 0) {
				notify({ text: "No results found", type: "error", time: 2 });
			}
		} catch (err) {
			console.error("geocode: Failed!", err);
			notify({ text: "Geocoding failed!", type: "error", time: 2 });
		}
		geocodeRunning = false;
	}

	function useGeocodeResult(r: GeocodeResult) {
		lat = Number(r.lat);
		lon = Number(r.lon);
		geocodeResults = [];
	}

	async function addScreen() {
		if (!newScreenName) return;
		try {
			const resp = await req.post<CinemaScreen>(
				`/source/${source.id}/cinema/screen`,
				{ name: newScreenName },
			);
			screens.push(resp);
			newScreenName = "";
		} catch (err) {
			console.error("addScreen: Failed!", err);
			notify({ text: "Failed adding screen!", type: "error", time: 2 });
		}
	}

	async function deleteScreen(screen: CinemaScreen) {
		try {
			await req.delete(`/source/${source.id}/cinema/screen/${screen.id}`);
			screens = screens.filter((s) => s.id !== screen.id);
		} catch (err) {
			console.error("deleteScreen: Failed!", err);
			notify({ text: "Failed deleting screen!", type: "error", time: 2 });
		}
	}

	async function save() {
		submitDisabled = true;
		const nid = notify({ text: "Saving Cinema Details", type: "loading" });
		try {
			await req.put(`/source/${source.id}/cinema`, {
				city,
				address,
				lat,
				lon,
				note,
				ratingOverall,
				ratingSnacks,
				ratingTech,
				ratingComfort,
			} as CinemaDetailsUpdateRequest);
			if (source.cinema) {
				source.cinema.city = city;
				source.cinema.address = address;
				source.cinema.lat = lat;
				source.cinema.lon = lon;
				source.cinema.note = note;
				source.cinema.ratingOverall = ratingOverall;
				source.cinema.ratingSnacks = ratingSnacks;
				source.cinema.ratingTech = ratingTech;
				source.cinema.ratingComfort = ratingComfort;
				source.cinema.screens = screens;
			}
			notify({ id: nid, text: "Cinema Details Saved!", type: "success" });
			onClose();
		} catch (err) {
			console.error("save cinema details: Failed!", err);
			notify({ id: nid, text: "Failed!", type: "error", time: 1 });
			error = "Failed!";
		}
		submitDisabled = false;
	}
</script>

<div class="wrap">
	<Modal
		title={source.name}
		desc="Details for this cinema"
		maxWidth="550px"
		{onClose}
		{error}
	>
		<SettingsList>
			<Setting title="City" desc="Where is this cinema?">
				<input type="text" name="city" placeholder="City" bind:value={city} />
			</Setting>
			<Setting title="Address">
				<input
					type="text"
					name="address"
					placeholder="Address"
					bind:value={address}
				/>
			</Setting>
			<Setting
				title="Coordinates"
				desc="Used for the map. Search fills them in via OpenStreetMap."
			>
				<div class="coords">
					<input type="number" placeholder="Latitude" bind:value={lat} />
					<input type="number" placeholder="Longitude" bind:value={lon} />
					<button onclick={() => geocode()} disabled={geocodeRunning}>
						Search
					</button>
				</div>
				{#if geocodeResults.length > 0}
					<div class="geocode-results">
						{#each geocodeResults as r (r.display_name)}
							<button class="plain" onclick={() => useGeocodeResult(r)}>
								{r.display_name}
							</button>
						{/each}
					</div>
				{/if}
			</Setting>
			<Setting title="Overall Rating" desc="Out of 10. Optional, like all ratings." row>
				<input
					class="rating-input"
					type="number"
					min="1"
					max="10"
					step="0.5"
					placeholder="-"
					bind:value={ratingOverall}
				/>
			</Setting>
			<button
				class="plain toggle-extra"
				onclick={() => (showExtraRatings = !showExtraRatings)}
			>
				{showExtraRatings ? "Hide" : "Show"} more ratings
			</button>
			{#if showExtraRatings}
				<Setting title="Popcorn & Snacks" row>
					<input
						class="rating-input"
						type="number"
						min="1"
						max="10"
						step="0.5"
						placeholder="-"
						bind:value={ratingSnacks}
					/>
				</Setting>
				<Setting title="Picture & Sound" row>
					<input
						class="rating-input"
						type="number"
						min="1"
						max="10"
						step="0.5"
						placeholder="-"
						bind:value={ratingTech}
					/>
				</Setting>
				<Setting title="Comfort" row>
					<input
						class="rating-input"
						type="number"
						min="1"
						max="10"
						step="0.5"
						placeholder="-"
						bind:value={ratingComfort}
					/>
				</Setting>
			{/if}
			<Setting title="Notes" desc="Anything to remember about this cinema.">
				<textarea
					name="note"
					placeholder="eg has no popcorn!"
					rows="2"
					bind:value={note}
				></textarea>
			</Setting>
			<Setting title="Screens" desc="Screens a watch can reference.">
				<div class="screens">
					{#each screens as screen (screen.id)}
						<div class="screen">
							<span>{screen.name}</span>
							<button class="plain" onclick={() => deleteScreen(screen)}>
								x
							</button>
						</div>
					{/each}
					<div class="screen-add">
						<input
							type="text"
							placeholder="Screen name"
							bind:value={newScreenName}
						/>
						<button onclick={() => addScreen()}>Add</button>
					</div>
				</div>
			</Setting>
			<button
				class="save-btn"
				onclick={() => save()}
				disabled={submitDisabled}
			>
				Save Details
			</button>
		</SettingsList>
	</Modal>
</div>

<style lang="scss">
	.wrap {
		color: $text-color;
	}

	.coords {
		display: flex;
		gap: 8px;

		input {
			min-width: 0;
		}

		button {
			width: max-content;
		}
	}

	.geocode-results {
		display: flex;
		flex-flow: column;
		gap: 4px;
		margin-top: 8px;

		button {
			text-align: left;
			font-size: 13px;
		}
	}

	.rating-input {
		max-width: 80px;
	}

	.toggle-extra {
		width: max-content;
		font-size: 13px;
	}

	.screens {
		display: flex;
		flex-flow: column;
		gap: 6px;

		.screen {
			display: flex;
			align-items: center;
			justify-content: space-between;

			button {
				width: max-content;
			}
		}

		.screen-add {
			display: flex;
			gap: 8px;

			button {
				width: max-content;
			}
		}
	}

	.save-btn {
		width: max-content;
		margin-left: auto;
	}
</style>
