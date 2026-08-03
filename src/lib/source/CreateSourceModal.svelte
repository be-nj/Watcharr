<script lang="ts">
	import Modal from "../Modal.svelte";
	import Setting from "../settings/Setting.svelte";
	import SettingsList from "../settings/SettingsList.svelte";
	import DropDown from "../DropDown.svelte";
	import { notify } from "../util/notify";
	import { req } from "../util/api";
	import type {
		GeocodeResult,
		WatchSource,
		WatchSourceAddRequest,
		WatchSourceType,
	} from "@/types";

	interface Props {
		onClose: (created?: WatchSource) => void;
	}

	let { onClose }: Props = $props();

	const sourceTypes = [
		{ id: "CINEMA", value: "Cinema" },
		{ id: "STREAMING", value: "Streaming" },
		{ id: "TV", value: "TV" },
		{ id: "SELFHOSTED", value: "Selfhosted" },
		{ id: "DISC", value: "Disc" },
		{ id: "OTHER", value: "Other" },
	];

	let sourceName = $state("");
	let sourceType: string | number | undefined = $state(undefined);
	let error = $state("");
	let submitDisabled = $state(false);

	// Cinemas are anchored on OpenStreetMap: search + pick instead of
	// typing (free form fallback below for unmapped cinemas).
	let isCinema = $derived(sourceType === "CINEMA");
	let osmQuery = $state("");
	let osmResults = $state<GeocodeResult[]>([]);
	let osmSearching = $state(false);
	let osmPick = $state<GeocodeResult | undefined>(undefined);
	let freeForm = $state(false);

	async function searchOsm() {
		if (!osmQuery) return;
		osmSearching = true;
		osmPick = undefined;
		try {
			osmResults = await req.get<GeocodeResult[]>(
				`/geocode?q=${encodeURIComponent(osmQuery)}`,
			);
			// Cinemas first.
			osmResults.sort((a, b) =>
				a.type === "cinema" === (b.type === "cinema")
					? 0
					: a.type === "cinema"
						? -1
						: 1,
			);
			if (osmResults.length === 0) {
				notify({ text: "No results found", type: "error", time: 2500 });
			}
		} catch (err) {
			console.error("searchOsm: Failed!", err);
			notify({ text: "Search failed!", type: "error", time: 2500 });
		}
		osmSearching = false;
	}

	function pickOsm(r: GeocodeResult) {
		osmPick = r;
		osmResults = [];
		if (!sourceName) {
			// First segment of the display name is usually the poi name.
			sourceName = r.display_name.split(",")[0];
		}
	}

	async function submitClicked() {
		if (!sourceName) {
			error = "Source must have a name!";
			return;
		}
		if (!sourceType) {
			error = "Source must have a type!";
			return;
		}
		if (isCinema && !osmPick && !freeForm) {
			error = "Pick the cinema from the search (or switch to free form).";
			return;
		}
		submitDisabled = true;
		const nid = notify({ text: "Creating Source", type: "loading" });
		try {
			const body: WatchSourceAddRequest = {
				name: sourceName,
				type: sourceType as WatchSourceType,
			};
			if (isCinema && osmPick) {
				body.osmType = osmPick.osm_type;
				body.osmId = osmPick.osm_id;
				body.wikidataId = osmPick.extratags?.wikidata ?? "";
				body.lat = Number(osmPick.lat);
				body.lon = Number(osmPick.lon);
				body.city =
					osmPick.address?.city ??
					osmPick.address?.town ??
					osmPick.address?.village ??
					"";
				body.address = [osmPick.address?.road, osmPick.address?.house_number]
					.filter(Boolean)
					.join(" ");
			}
			const resp = await req.post<WatchSource>("/source", body);
			notify({ id: nid, text: "Source Created!", type: "success" });
			onClose(resp);
		} catch (err) {
			console.error("addSource: Failed!", err);
			notify({ id: nid, text: "Failed!", type: "error", time: 2500 });
			error = "Failed!";
		}
		submitDisabled = false;
	}
</script>

<div class="wrap">
	<Modal
		title="Create A Watch Source"
		desc="Sources are shared with everyone on this server"
		maxWidth="500px"
		onClose={() => onClose()}
		{error}
	>
		<SettingsList>
			<Setting
				title="Type"
				desc="What kind of source is this? Cinemas are picked from OpenStreetMap."
			>
				<DropDown
					options={sourceTypes}
					isDropDownItem={true}
					bind:active={sourceType}
					placeholder="Type"
				/>
			</Setting>
			{#if isCinema && !freeForm}
				<Setting
					title="Find Cinema"
					desc="Search OpenStreetMap and pick your cinema."
				>
					<div class="osm-search">
						<input
							type="text"
							placeholder="eg CineStar Metropolis Frankfurt"
							bind:value={osmQuery}
							onkeydown={(ev) => ev.key === "Enter" && searchOsm()}
						/>
						<button onclick={() => searchOsm()} disabled={osmSearching}>
							Search
						</button>
					</div>
					{#if osmResults.length > 0}
						<div class="osm-results">
							{#each osmResults as r (`${r.osm_type}${r.osm_id}`)}
								<button class="plain" onclick={() => pickOsm(r)}>
									{r.type === "cinema" ? "🎬 " : ""}{r.display_name}
								</button>
							{/each}
						</div>
					{/if}
					{#if osmPick}
						<p class="osm-picked">Picked: {osmPick.display_name}</p>
					{/if}
					<button class="plain free-form-toggle" onclick={() => (freeForm = true)}>
						Cinema not on OpenStreetMap? Create free form
					</button>
				</Setting>
			{/if}
			<Setting title="Name" desc="What should we call this source?">
				<input
					type="text"
					name="name"
					placeholder="Name"
					bind:value={sourceName}
				/>
			</Setting>
			<button
				class="add-source-btn"
				onclick={() => submitClicked()}
				disabled={submitDisabled}
			>
				Create Source
			</button>
		</SettingsList>
	</Modal>
</div>

<style lang="scss">
	.add-source-btn {
		width: max-content;
		margin-left: auto;
	}

	.wrap {
		color: $text-color;
	}

	.osm-search {
		display: flex;
		gap: 8px;

		input {
			min-width: 0;
		}

		button {
			width: max-content;
		}
	}

	.osm-results {
		display: flex;
		flex-flow: column;
		gap: 4px;
		margin-top: 8px;

		button {
			text-align: left;
			font-size: 13px;
		}
	}

	.osm-picked {
		margin-top: 8px;
		font-size: 13px;
	}

	.free-form-toggle {
		margin-top: 8px;
		font-size: 13px;
		width: max-content;
	}
</style>
