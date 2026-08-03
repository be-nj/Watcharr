<script lang="ts">
	import { resolve } from "$app/paths";
	import Error from "@/lib/Error.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import { req } from "@/lib/util/api";
	import type { CinemaStatsResponse } from "@/types";

	let statsPromise = req.get<CinemaStatsResponse>("/stats/cinema");

	function maxVisits(stats: CinemaStatsResponse) {
		return Math.max(...stats.years.map((y) => y.visits), 1);
	}
</script>

<svelte:head>
	<title>Cinema Stats</title>
</svelte:head>

<div class="content">
	<div class="inner">
		<a href={resolve("/sources")} class="back">← All Sources</a>
		<h2>Cinema Stats</h2>
		{#await statsPromise}
			<Spinner />
		{:then stats}
			{#if stats.totalVisits === 0}
				<p class="empty">
					No cinema visits yet. Attach a cinema to your watches from the
					activity list (or use Log Cinema Visit on a movie page).
				</p>
			{:else}
				<div class="totals">
					<div class="total">
						<span class="num">{stats.totalVisits}</span>
						<span>{stats.totalVisits === 1 ? "visit" : "visits"}</span>
					</div>
					<div class="total">
						<span class="num">{stats.distinctCinemas}</span>
						<span>{stats.distinctCinemas === 1 ? "cinema" : "cinemas"}</span>
					</div>
				</div>

				<h3>Visits Per Year</h3>
				<div class="years">
					{#each stats.years as y (y.year)}
						<div class="year">
							<span class="label">{y.year}</span>
							<div class="bar-wrap">
								<div
									class="bar"
									style="width: {(y.visits / maxVisits(stats)) * 100}%"
								></div>
								<span class="count">{y.visits}</span>
							</div>
							<span class="snacks">
								{#if y.ratingSnacksAvg}
									🍿 {y.ratingSnacksAvg.toFixed(1)}
								{/if}
							</span>
						</div>
					{/each}
				</div>

				<h3>Your Cinemas</h3>
				<div class="ranking">
					{#each stats.ranking as r, i (r.id)}
						<a href={resolve(`/sources/${r.id}`)} class="cinema">
							<span class="pos">{i + 1}.</span>
							<span class="name">
								{r.name}
								{#if r.city}
									<span class="city">{r.city}</span>
								{/if}
							</span>
							<span class="visits">
								{r.visits}
								{r.visits === 1 ? "visit" : "visits"}
							</span>
							<span class="rating">
								{#if r.ratingOverallAvg}
									Ø {r.ratingOverallAvg.toFixed(1)}/10
								{/if}
							</span>
						</a>
					{/each}
				</div>
			{/if}
		{:catch err}
			<Error error={err} pretty="Failed to load cinema stats!" />
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

	h3 {
		margin: 20px 0 10px 0;
	}

	.empty {
		font-style: italic;
		margin-bottom: 10px;
	}

	.totals {
		display: flex;
		gap: 15px;

		.total {
			display: flex;
			flex-flow: column;
			align-items: center;
			padding: 15px 25px;
			background-color: var(--accent-color);
			border-radius: 10px;

			.num {
				font-size: 26px;
				font-weight: bold;
				font-family:
					sans-serif,
					system-ui,
					-apple-system,
					BlinkMacSystemFont;
			}
		}
	}

	.years {
		display: flex;
		flex-flow: column;
		gap: 6px;

		.year {
			display: flex;
			align-items: center;
			gap: 10px;

			.label {
				width: 45px;
				flex-shrink: 0;
			}

			.bar-wrap {
				display: flex;
				align-items: center;
				gap: 8px;
				flex: 1;

				.bar {
					height: 18px;
					min-width: 4px;
					background-color: var(--accent-color-hover);
					border-radius: 4px;
				}
			}

			.snacks {
				width: 65px;
				flex-shrink: 0;
				text-align: right;
			}
		}
	}

	.ranking {
		display: flex;
		flex-flow: column;
		gap: 8px;

		.cinema {
			display: flex;
			align-items: center;
			gap: 10px;
			padding: 12px 15px;
			background-color: var(--accent-color);
			border-radius: 10px;
			text-decoration: none;

			&:hover {
				background-color: var(--accent-color-hover);
			}

			.pos {
				width: 25px;
				flex-shrink: 0;
			}

			.name {
				flex: 1;
				display: flex;
				flex-flow: column;

				.city {
					font-size: 13px;
					opacity: 0.7;
				}
			}

			.visits,
			.rating {
				flex-shrink: 0;
			}
		}
	}
</style>
