<script lang="ts">
	import { updateActivity, removeActivity, req } from "@/lib/util/api";
	import Modal from "./Modal.svelte";
	import DropDown from "./DropDown.svelte";
	import type {
		Activity,
		ActivityDetails,
		ActivityDetailsUpdateRequest,
		Tag,
		WatchSource,
	} from "@/types";
	import { store } from "@/store.svelte";
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

	async function update() {
		const dateObj = validateNewDate();
		if (dateObj && isDateTimeValid && isDateTimeChanged) {
			const updatedActivity = await updateActivity(activity, dateObj);
			if (!updatedActivity) {
				// Failed..
				return;
			}
			onUpdated(updatedActivity.id, updatedActivity);
			onClose();
			return;
		}
		notify({ text: "Unable to try updating!", type: "error" });
		console.error(
			"ActivityEditor: Can't try updating, data missing/invalid:",
			dateObj,
			isDateTimeValid,
			isDateTimeChanged,
		);
	}

	async function remove() {
		const success = await removeActivity(activity.id);
		if (!success) {
			return;
		}
		onRemoved(activity);
		onClose();
	}

	// Per watch details (source, language, tags, note).
	let showDetails = $state(false);
	let sources = $state<WatchSource[]>([]);
	let selectedSourceId: string | number | undefined = $state(
		activity.details?.watchSourceId,
	);
	let selectedScreenId: string | number | undefined = $state(
		activity.details?.cinemaScreenId,
	);
	let audioLang = $state(activity.details?.audioLang ?? "");
	let subtitleLang = $state(activity.details?.subtitleLang ?? "");
	let detailsNote = $state(activity.details?.note ?? "");
	let selectedTagIds = $state<number[]>(
		activity.details?.tags?.map((t) => t.id) ?? [],
	);
	let detailsSaving = $state(false);

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
	let allTags = $derived(store.tags);

	async function toggleDetails() {
		showDetails = !showDetails;
		if (showDetails && sources.length === 0) {
			try {
				sources = await req.get<WatchSource[]>("/source");
			} catch (err) {
				console.error("ActivityEditor: Failed getting sources!", err);
			}
		}
	}

	function toggleTag(tag: Tag) {
		if (selectedTagIds.includes(tag.id)) {
			selectedTagIds = selectedTagIds.filter((id) => id !== tag.id);
		} else {
			selectedTagIds = [...selectedTagIds, tag.id];
		}
	}

	async function saveDetails() {
		detailsSaving = true;
		const nid = notify({ text: "Saving Details", type: "loading" });
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
					tagIds: selectedTagIds,
				} as ActivityDetailsUpdateRequest,
			);
			activity.details = resp;
			notify({ id: nid, text: "Details Saved!", type: "success" });
			onUpdated(activity.id, activity);
		} catch (err) {
			console.error("ActivityEditor: Failed saving details!", err);
			notify({ id: nid, text: "Failed!", type: "error", time: 1 });
		}
		detailsSaving = false;
	}
</script>

<Modal title="Edit Activity" desc={activityMessage} maxWidth="400px" {onClose}>
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

		<button class="plain toggle-details" onclick={toggleDetails}>
			{showDetails ? "Hide" : "Show"} watch details
		</button>
		{#if showDetails}
			<h3>Watched Via</h3>
			<DropDown
				options={sourceOptions}
				isDropDownItem={true}
				bind:active={selectedSourceId}
				placeholder="Source"
			/>
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
			{#if allTags?.length > 0}
				<h3>Tags</h3>
				<div class="tags">
					{#each allTags as tag (tag.id)}
						<button
							class="plain tag"
							class:selected={selectedTagIds.includes(tag.id)}
							style="color: {tag.color}; background-color: {tag.bgColor};"
							onclick={() => toggleTag(tag)}
						>
							{tag.name}
						</button>
					{/each}
				</div>
			{/if}
			<h3>Note</h3>
			<textarea
				placeholder="Anything to remember about this watch"
				rows="2"
				bind:value={detailsNote}
			></textarea>
			<button
				class="save-details"
				onclick={saveDetails}
				disabled={detailsSaving}
			>
				Save Details
			</button>
		{/if}

		<div class="button-row">
			<button class="danger" onclick={remove}>Delete</button>
			<div>
				<button
					onclick={update}
					disabled={!(isDateTimeChanged && isDateTimeValid)}>Update</button
				>
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

		.toggle-details {
			width: max-content;
			font-size: 13px;
		}

		.langs {
			display: flex;
			gap: 8px;

			input {
				min-width: 0;
			}
		}

		.tags {
			display: flex;
			flex-flow: wrap;
			gap: 6px;

			.tag {
				width: max-content;
				padding: 2px 10px;
				border-radius: 10px;
				font-size: 13px;
				opacity: 0.5;

				&.selected {
					opacity: 1;
				}
			}
		}

		.save-details {
			width: max-content;
			margin-left: auto;
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
