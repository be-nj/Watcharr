<script lang="ts">
	import { resolve } from "$app/paths";
	import DropDown from "@/lib/DropDown.svelte";
	import Error from "@/lib/Error.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import { req } from "@/lib/util/api";
	import type { CinemaStatsResponse, WatchSource } from "@/types";
	import { onDestroy, onMount } from "svelte";
	import "leaflet/dist/leaflet.css";

	let mapElement: HTMLDivElement | undefined = $state();
	let map: import("leaflet").Map | undefined;
	let markerLayer: import("leaflet").LayerGroup | undefined;
	let leaflet: typeof import("leaflet") | undefined;
	let cinemas: WatchSource[] = [];
	let visitsById: Map<number, number> = new Map();
	let cinemasWithCoords = $state(0);
	let shownCinemas = $state(0);
	let ratedFilter: string | undefined = $state("All");
	let loadPromise = $state<Promise<void> | undefined>(undefined);

	async function load() {
		// Leaflet greift auf window zu, deshalb erst im Browser importieren.
		leaflet = (await import("leaflet")).default;
		const L = leaflet;
		const [sources, stats] = await Promise.all([
			req.get<WatchSource[]>("/source"),
			req.get<CinemaStatsResponse>("/stats/cinema"),
		]);
		cinemas = sources.filter(
			(s) =>
				s.type === "CINEMA" &&
				s.cinema?.lat !== undefined &&
				s.cinema?.lat !== null &&
				s.cinema?.lon !== undefined &&
				s.cinema?.lon !== null,
		);
		visitsById = new Map(stats.ranking.map((r) => [r.id, r.visits]));
		cinemasWithCoords = cinemas.length;
		if (!mapElement) {
			return;
		}
		map = L.map(mapElement);
		L.tileLayer("https://tile.openstreetmap.org/{z}/{x}/{y}.png", {
			maxZoom: 19,
			attribution:
				'&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
		}).addTo(map);
		markerLayer = L.layerGroup().addTo(map);
		renderMarkers(true);
	}

	// (Re)draws the cinema pins: size follows the own visit count,
	// rated cinemas are filled, and the filter can limit to
	// rated/unrated ones.
	function renderMarkers(fitView = false) {
		const L = leaflet;
		if (!L || !map || !markerLayer) {
			return;
		}
		markerLayer.clearLayers();
		const shown = cinemas.filter((c) => {
			if (ratedFilter === "Rated") {
				return !!c.ratingAverage;
			}
			if (ratedFilter === "Unrated") {
				return !c.ratingAverage;
			}
			return true;
		});
		shownCinemas = shown.length;
		const bounds = L.latLngBounds([]);
		for (const c of shown) {
			const pos: [number, number] = [c.cinema!.lat!, c.cinema!.lon!];
			bounds.extend(pos);
			const visits = visitsById.get(c.id) ?? 0;
			const rating = c.ratingAverage
				? `<br/>&Oslash; ${c.ratingAverage.toFixed(1)}/10 (${c.ratingCount} rated ${c.ratingCount === 1 ? "visit" : "visits"})`
				: "";
			const visitsLine =
				visits > 0
					? `<br/>${visits} ${visits === 1 ? "visit" : "visits"}`
					: "<br/>Not visited yet";
			L.circleMarker(pos, {
				radius: 6 + Math.min(Math.sqrt(visits) * 3, 14),
				color: "#3388ff",
				weight: 2,
				fillColor: "#3388ff",
				fillOpacity: c.ratingAverage ? 0.7 : 0.15,
			})
				.addTo(markerLayer)
				.bindPopup(
					`<b><a href="${resolve(`/sources/${c.id}`)}">${c.name}</a></b>` +
						`${c.cinema?.city ? `<br/>${c.cinema.city}` : ""}${visitsLine}${rating}`,
				);
		}
		if (fitView) {
			if (shown.length > 0) {
				map.fitBounds(bounds, { padding: [40, 40] });
			} else {
				map.setView([51.16, 10.45], 6); // Deutschland
			}
		}
	}

	onMount(() => {
		loadPromise = load();
	});

	onDestroy(() => {
		map?.remove();
	});
</script>

<svelte:head>
	<title>Cinema Map</title>
</svelte:head>

<div class="content">
	<div class="inner">
		<a href={resolve("/sources")} class="back">← All Sources</a>
		<h2>Cinema Map</h2>
		{#if loadPromise}
			{#await loadPromise}
				<Spinner />
			{:then}
				<div class="controls">
					<DropDown
						options={["All", "Rated", "Unrated"]}
						bind:active={ratedFilter}
						placeholder="Filter"
						onChange={() => renderMarkers()}
					/>
					<span class="shown">
						{shownCinemas} of {cinemasWithCoords}
						{cinemasWithCoords === 1 ? "cinema" : "cinemas"} shown (pin size =
						your visits)
					</span>
				</div>
				{#if cinemasWithCoords === 0}
					<p class="empty">
						No cinemas with coordinates yet. Add them on the source pages.
					</p>
				{/if}
			{:catch err}
				<Error error={err} pretty="Failed to load map!" />
			{/await}
		{/if}
		<div class="map" bind:this={mapElement}></div>
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
			width: 100%;
			max-width: 1100px;
		}
	}

	.back {
		margin-bottom: 10px;
		width: max-content;
	}

	h2 {
		margin-bottom: 15px;
	}

	.controls {
		display: flex;
		align-items: center;
		gap: 12px;
		margin-bottom: 10px;
		z-index: 2;

		.shown {
			font-size: 13px;
			opacity: 0.8;
		}
	}

	.empty {
		font-style: italic;
		margin-bottom: 10px;
	}

	.map {
		width: 100%;
		height: 70vh;
		border-radius: 10px;
		z-index: 1;
	}
</style>
