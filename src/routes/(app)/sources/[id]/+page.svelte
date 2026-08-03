<script lang="ts">
	import { goto } from "$app/navigation";
	import { resolve } from "$app/paths";
	import { page } from "$app/state";
	import DropDown from "@/lib/DropDown.svelte";
	import Error from "@/lib/Error.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import Setting from "@/lib/settings/Setting.svelte";
	import SettingsList from "@/lib/settings/SettingsList.svelte";
	import { req } from "@/lib/util/api";
	import { notify } from "@/lib/util/notify";
	import type {
		CinemaDetailsUpdateRequest,
		CinemaScreen,
		GeocodeResult,
		WatchSource,
		WatchSourceType,
		WatchSourceRating,
		WatchSourceWatch,
	} from "@/types";

	const sourceTypes = [
		{ id: "CINEMA", value: "Cinema" },
		{ id: "STREAMING", value: "Streaming" },
		{ id: "TV", value: "TV" },
		{ id: "SELFHOSTED", value: "Selfhosted" },
		{ id: "DISC", value: "Disc" },
		{ id: "OTHER", value: "Other" },
	];

	let sourceId = $derived(Number(page.params.id));
	let source = $state<WatchSource | undefined>(undefined);
	let watches = $state<WatchSourceWatch[]>([]);
	let ratings = $state<WatchSourceRating[]>([]);
	let loadPromise = $state(load());

	// Editierbare Felder
	let name = $state("");
	let sourceType: string | number | undefined = $state(undefined);
	let city = $state("");
	let address = $state("");
	let lat = $state<number | undefined>(undefined);
	let lon = $state<number | undefined>(undefined);
	let note = $state("");
	let screens = $state<CinemaScreen[]>([]);
	let newScreenName = $state("");
	let geocodeResults = $state<GeocodeResult[]>([]);
	let geocodeRunning = $state(false);
	let saving = $state(false);

	let isCinema = $derived(sourceType === "CINEMA");

	async function load() {
		const s = await req.get<WatchSource>(`/source/${Number(page.params.id)}`);
		source = s;
		name = s.name;
		sourceType = s.type;
		city = s.cinema?.city ?? "";
		address = s.cinema?.address ?? "";
		lat = s.cinema?.lat;
		lon = s.cinema?.lon;
		note = s.cinema?.note ?? "";
		screens = s.cinema?.screens ?? [];
		watches = await req.get<WatchSourceWatch[]>(
			`/source/${Number(page.params.id)}/watches`,
		);
		if (s.type === "CINEMA") {
			ratings = await req.get<WatchSourceRating[]>(
				`/source/${Number(page.params.id)}/ratings`,
			);
		}
	}

	async function save() {
		if (!name) {
			notify({ text: "Source must have a name!", type: "error", time: 2 });
			return;
		}
		saving = true;
		const nid = notify({ text: "Saving Source", type: "loading" });
		try {
			await req.put(`/source/${sourceId}`, {
				name,
				type: sourceType as WatchSourceType,
			});
			if (isCinema) {
				await req.put(`/source/${sourceId}/cinema`, {
					city,
					address,
					lat,
					lon,
					note,
				} as CinemaDetailsUpdateRequest);
			}
			notify({ id: nid, text: "Source Saved!", type: "success" });
		} catch (err) {
			console.error("save source: Failed!", err);
			notify({ id: nid, text: "Failed!", type: "error", time: 1 });
		}
		saving = false;
	}

	async function geocode() {
		const query = [name, city].filter(Boolean).join(", ");
		geocodeRunning = true;
		try {
			geocodeResults = await req.get<GeocodeResult[]>(
				`/geocode?q=${encodeURIComponent(query)}`,
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
				`/source/${sourceId}/cinema/screen`,
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
			await req.delete(`/source/${sourceId}/cinema/screen/${screen.id}`);
			screens = screens.filter((s) => s.id !== screen.id);
		} catch (err) {
			console.error("deleteScreen: Failed!", err);
			notify({ text: "Failed deleting screen!", type: "error", time: 2 });
		}
	}

	async function deleteSource() {
		const nid = notify({ text: "Deleting Source", type: "loading" });
		try {
			await req.delete(`/source/${sourceId}`);
			notify({ id: nid, text: "Source Deleted!", type: "success" });
			goto(resolve("/sources"));
		} catch (err) {
			console.error("deleteSource: Failed!", err);
			notify({ id: nid, text: "Failed!", type: "error", time: 1 });
		}
	}

	function watchLink(w: WatchSourceWatch) {
		return resolve(`/${w.type === "tv" ? "tv" : "movie"}/${w.tmdbId}`);
	}

	function formatDate(d: string) {
		return new Date(Date.parse(d)).toLocaleDateString();
	}
</script>

<svelte:head>
	<title>{source?.name ?? "Watch Source"}</title>
</svelte:head>

<div class="content">
	<div class="inner">
		<a href={resolve("/sources")} class="back">← All Sources</a>
		{#await loadPromise}
			<Spinner />
		{:then}
			<h2>{source?.name}</h2>
			<SettingsList>
				<Setting title="Name">
					<input type="text" name="name" placeholder="Name" bind:value={name} />
				</Setting>
				<Setting title="Type">
					<DropDown
						options={sourceTypes}
						isDropDownItem={true}
						bind:active={sourceType}
						placeholder="Type"
					/>
				</Setting>
				{#if isCinema}
					<Setting title="City">
						<input type="text" placeholder="City" bind:value={city} />
					</Setting>
					<Setting title="Address">
						<input type="text" placeholder="Address" bind:value={address} />
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
					{#if source?.cinema?.osmId}
						<Setting
							title="OpenStreetMap"
							desc="This cinema is anchored on OSM; local data is a cached copy."
						>
							<a
								href={`https://www.openstreetmap.org/${source.cinema.osmType}/${source.cinema.osmId}`}
								target="_blank"
								rel="noopener noreferrer"
							>
								{source.cinema.osmType}/{source.cinema.osmId} ↗
							</a>
						</Setting>
					{/if}
					<Setting
						title="Visit Ratings"
						desc="Ratings are given per visit when logging a watch."
					>
						{#if source?.ratingCount}
							<p class="rating-aggregate">
								Ø {source.ratingAverage?.toFixed(1)}/10 from {source.ratingCount}
								rated {source.ratingCount === 1 ? "visit" : "visits"}
							</p>
						{:else}
							<p class="empty">No visit ratings yet.</p>
						{/if}
						{#if ratings.length > 0}
							<div class="rating-list">
								{#each ratings as r (r.date + (r.username || "anon"))}
									<div class="rating-entry" class:own={r.own}>
										<span class="who">
											{r.username || "Anonymous"}{r.own ? " (you)" : ""}
											· {formatDate(r.date)}
										</span>
										<span class="dims">
											{r.ratingOverall}/10
											{r.ratingSnacks ? ` · Popcorn ${r.ratingSnacks}` : ""}
											{r.ratingTech ? ` · Tech ${r.ratingTech}` : ""}
											{r.ratingComfort ? ` · Comfort ${r.ratingComfort}` : ""}
										</span>
									</div>
								{/each}
							</div>
						{/if}
					</Setting>
					<Setting title="Notes" desc="Anything to remember about this cinema.">
						<textarea
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
				{/if}
				<div class="button-row">
					<button class="danger" onclick={() => deleteSource()}>Delete</button>
					<button onclick={() => save()} disabled={saving}>Save</button>
				</div>
			</SettingsList>

			<h3>Watches</h3>
			{#if watches.length === 0}
				<p class="empty">No watches used this source yet.</p>
			{:else}
				<div class="watches">
					{#each watches as w (w.activityId)}
						<a href={watchLink(w)} class="watch">
							<span class="title">{w.title}</span>
							<span class="meta">
								{formatDate(w.date)}{w.screenName ? ` · ${w.screenName}` : ""}
							</span>
						</a>
					{/each}
				</div>
			{/if}
		{:catch err}
			<Error error={err} pretty="Failed to load source!" />
		{/await}
	</div>
</div>

<style lang="scss">
	.content {
		display: flex;
		width: 100%;
		justify-content: center;
		padding: 0 30px 30px 30px;

		.inner {
			display: flex;
			flex-flow: column;
			min-width: 400px;
			max-width: 700px;
			width: 100%;

			@media screen and (max-width: 420px) {
				min-width: 100%;
			}
		}
	}

	.back {
		margin-bottom: 10px;
		width: max-content;
	}

	h2 {
		margin-bottom: 15px;
	}

	h3 {
		margin: 25px 0 8px 0;
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

	.rating-aggregate {
		font-size: 15px;
		font-weight: bold;
	}

	.rating-list {
		display: flex;
		flex-flow: column;
		gap: 6px;
		margin-top: 8px;

		.rating-entry {
			display: flex;
			flex-flow: column;
			padding: 6px 10px;
			border-radius: 8px;
			background-color: $accent-color;

			&.own {
				outline: 1px solid $text-color;
			}

			.who {
				font-size: 12px;
				opacity: 0.7;
			}

			.dims {
				font-size: 14px;
			}
		}
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

	.button-row {
		display: flex;
		justify-content: space-between;
		margin-top: 10px;

		button {
			width: max-content;
		}
	}

	.empty {
		font-style: italic;
	}

	.watches {
		display: flex;
		flex-flow: column;
		gap: 6px;

		.watch {
			display: flex;
			align-items: center;
			justify-content: space-between;
			padding: 8px 12px;
			border-radius: 8px;
			background-color: $accent-color;
			text-decoration: none;

			.meta {
				font-size: 13px;
				opacity: 0.7;
			}
		}
	}
</style>
