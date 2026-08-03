<script lang="ts">
	import { updateActivity, removeActivity, req } from "@/lib/util/api";
	import Modal from "./Modal.svelte";
	import DropDown from "./DropDown.svelte";
	import type {
		Activity,
		ActivityDetails,
		ActivityDetailsUpdateRequest,
		WatchSource,
	} from "@/types";
	import { onMount } from "svelte";
	import { resolve } from "$app/paths";
	import { notify } from "./util/notify";

	interface Props {
		activity: Activity;
		activityMessage: string;
		onClose: () => void;
		onRemoved: (activity: Activity) => void;
		onUpdated: (activityId: number, updatedActivity: Activity) => void;
	}

	let { activity, activityMessage, onClose, onRemoved, onUpdated }: Props =
		$props();

	let isDateTimeValid = $state(true);
	let currentDateObject = new Date(
		Date.parse(activity.customDate ?? activity.createdAt),
	);
	let currentDateString = dateToInputDateString(currentDateObject);
	let currentTimeString = dateToInputTimeString(currentDateObject);
	let selectedDateString = $state(currentDateString);
	let selectedTimeString = $state(currentTimeString);
	let isDateTimeChanged: boolean = $derived(
		currentDateString != selectedDateString ||
			currentTimeString != selectedTimeString,
	);

	function dateToInputDateString(date: Date) {
		const year = date.getFullYear();
		const month = (date.getMonth() + 1).toString().padStart(2, "0");
		const day = date.getDate().toString().padStart(2, "0");
		return `${year}-${month}-${day}`;
	}

	function dateToInputTimeString(date: Date) {
		const hours = date.getHours().toString().padStart(2, "0");
		const minutes = date.getMinutes().toString().padStart(2, "0");
		return `${hours}:${minutes}`;
	}

	function validateNewDate() {
		try {
			const epochMillis = Date.parse(
				`${selectedDateString} ${selectedTimeString}`,
			);
			const dateObj = new Date(epochMillis);
			if (isNaN(dateObj.getTime())) {
				isDateTimeValid = false;
				return;
			}
			isDateTimeValid = true;
			return dateObj;
		} catch (err) {
			console.error("ActivityEditor: validateNewDate failed!", err);
			isDateTimeValid = false;
		}
	}

	// Saves date (when changed) and watch details together.
	async function save() {
		saving = true;
		try {
			if (isDateTimeChanged) {
				const dateObj = validateNewDate();
				if (!dateObj || !isDateTimeValid) {
					notify({ text: "Invalid date/time!", type: "error" });
					return;
				}
				const updatedActivity = await updateActivity(activity, dateObj);
				if (!updatedActivity) {
					return;
				}
				activity.customDate = updatedActivity.customDate;
			}
			if (!(await saveDetails())) {
				return;
			}
			onUpdated(activity.id, activity);
			onClose();
		} finally {
			saving = false;
		}
	}

	async function remove() {
		const success = await removeActivity(activity.id);
		if (!success) {
			return;
		}
		onRemoved(activity);
		onClose();
	}

	// Per watch details (source, language, note).
	let sources = $state<WatchSource[]>([]);
	let saving = $state(false);
	let selectedSourceId: string | number | undefined = $state(
		activity.details?.watchSourceId,
	);
	let selectedScreenId: string | number | undefined = $state(
		activity.details?.cinemaScreenId,
	);
	let audioLang = $state(activity.details?.audioLang ?? "");
	let subtitleLang = $state(activity.details?.subtitleLang ?? "");
	let detailsNote = $state(activity.details?.note ?? "");

	let selectedSource = $derived(
		sources.find((s) => s.id === Number(selectedSourceId)),
	);
	let screenOptions = $derived(
		selectedSource?.cinema?.screens?.map((s) => ({
			id: s.id,
			value: s.name,
		})) ?? [],
	);
	let sourceOptions = $derived(
		sources.map((s) => ({ id: s.id, value: s.name })),
	);
	onMount(async () => {
		try {
			sources = await req.get<WatchSource[]>("/source");
		} catch (err) {
			console.error("ActivityEditor: Failed getting sources!", err);
		}
		// The watched list endpoints don't include activity details, so
		// fetch them fresh from the activity endpoint (which does).
		try {
			const acts = await req.get<Activity[]>(`/activity/${activity.watchedId}`);
			const fresh = acts.find((a) => a.id === activity.id);
			if (fresh?.details) {
				activity.details = fresh.details;
				selectedSourceId = fresh.details.watchSourceId;
				selectedScreenId = fresh.details.cinemaScreenId;
				audioLang = fresh.details.audioLang ?? "";
				subtitleLang = fresh.details.subtitleLang ?? "";
				detailsNote = fresh.details.note ?? "";
			}
		} catch (err) {
			console.error("ActivityEditor: Failed getting activity details!", err);
		}
	});

	async function saveDetails(): Promise<boolean> {
		try {
			const resp = await req.put<ActivityDetails>(
				`/activity/${activity.id}/details`,
				{
					watchSourceId: selectedSourceId
						? Number(selectedSourceId)
						: undefined,
					cinemaScreenId:
						selectedSourceId && selectedScreenId
							? Number(selectedScreenId)
							: undefined,
					audioLang,
					subtitleLang,
					note: detailsNote,
				} as ActivityDetailsUpdateRequest,
			);
			activity.details = resp;
			return true;
		} catch (err) {
			console.error("ActivityEditor: Failed saving details!", err);
			notify({ text: "Failed saving details!", type: "error", time: 2 });
			return false;
		}
	}
</script>

<Modal title="Edit Activity" desc={activityMessage} maxWidth="520px" {onClose}>
	<div class="centered">
		<h3>Date</h3>
		<input
			id="activity-date"
			type="date"
			bind:value={selectedDateString}
			onchange={validateNewDate}
			class:invalid={!isDateTimeValid}
		/>
		<h3>Time</h3>
		<input
			id="activity-time"
			type="time"
			bind:value={selectedTimeString}
			onchange={validateNewDate}
		/>

		<h3>Watched Via</h3>
			<DropDown
				options={sourceOptions}
				isDropDownItem={true}
				bind:active={selectedSourceId}
				placeholder="Source"
			/>
			{#if selectedSource?.cinema}
			<a class="cinema-link" href={resolve(`/sources/${selectedSource.id}`)}>
				View & rate {selectedSource.name}
				{selectedSource.cinema.ratingOverall
					? `(currently ${selectedSource.cinema.ratingOverall}/10)`
					: ""} →
			</a>
		{/if}
		{#if selectedSource?.cinema && screenOptions.length > 0}
				<h3>Screen</h3>
				<DropDown
					options={screenOptions}
					isDropDownItem={true}
					bind:active={selectedScreenId}
					placeholder="Screen"
				/>
			{/if}
			<h3>Language</h3>
			<div class="langs">
				<input
					type="text"
					placeholder="Audio (eg de, en)"
					maxlength="5"
					bind:value={audioLang}
				/>
				<input
					type="text"
					placeholder="Subtitles (none, de, ..)"
					maxlength="5"
					bind:value={subtitleLang}
				/>
			</div>
			<h3>Note</h3>
		<textarea
			placeholder="Anything to remember about this watch"
			rows="2"
			bind:value={detailsNote}
		></textarea>

		<div class="button-row">
			<button class="danger" onclick={remove}>Delete</button>
			<div>
				<button onclick={save} disabled={saving}>Save</button>
			</div>
		</div>
	</div>
</Modal>

<style lang="scss">
	.centered {
		display: flex;
		flex-flow: column;
		gap: 10px;
		height: 100%;

		.cinema-link {
			font-size: 13px;
			width: max-content;
		}

		.langs {
			display: flex;
			gap: 8px;

			input {
				min-width: 0;
			}
		}



		h3 {
			font-size: 16px;
			font-family:
				sans-serif,
				system-ui,
				-apple-system,
				BlinkMacSystemFont;
		}

		.button-row {
			display: flex;
			flex-flow: row;
			justify-content: space-between;
			margin-top: 10px;

			button {
				margin-top: auto;
				width: max-content;
			}
		}
	}
</style>
