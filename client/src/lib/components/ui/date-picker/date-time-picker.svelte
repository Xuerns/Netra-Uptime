<script lang="ts">
	import { CalendarDate, type DateValue } from '@internationalized/date';
	import * as Popover from '$lib/components/ui/popover/index.js';
	import { Calendar } from '$lib/components/ui/calendar/index.js';
	import { Calendar as CalendarIcon, Clock } from 'lucide-svelte';
	import { cn } from '$lib/utils';

	interface Props {
		value?: string;
		placeholder?: string;
		disabled?: boolean;
		class?: string;
		id?: string;
	}

	let {
		value = $bindable(''),
		placeholder = 'Select date & time',
		disabled = false,
		class: className,
		id
	}: Props = $props();

	let isOpen = $state(false);

	function parseValue(val: string) {
		if (!val) {
			return { date: undefined as CalendarDate | undefined, hour: '00', minute: '00' };
		}
		const match = val.match(/^(\d{4})-(\d{2})-(\d{2})(?:[T ](\d{2}):(\d{2}))?/);
		if (match) {
			const [, y, m, d, h, min] = match;
			return {
				date: new CalendarDate(parseInt(y, 10), parseInt(m, 10), parseInt(d, 10)),
				hour: (h ?? '00').padStart(2, '0'),
				minute: (min ?? '00').padStart(2, '0')
			};
		}
		const d = new Date(val);
		if (!isNaN(d.getTime())) {
			return {
				date: new CalendarDate(d.getFullYear(), d.getMonth() + 1, d.getDate()),
				hour: d.getHours().toString().padStart(2, '0'),
				minute: d.getMinutes().toString().padStart(2, '0')
			};
		}
		return { date: undefined as CalendarDate | undefined, hour: '00', minute: '00' };
	}

	let selectedDate = $state<DateValue | undefined>(undefined);
	let hour = $state('00');
	let minute = $state('00');

	// Synchronize internal state whenever external value changes
	$effect(() => {
		const p = parseValue(value);
		selectedDate = p.date;
		hour = p.hour;
		minute = p.minute;
	});

	function updateValue(date: DateValue | undefined, h: string, m: string) {
		if (!date) {
			value = '';
			return;
		}
		const pad = (n: number | string) => n.toString().padStart(2, '0');
		value = `${date.year}-${pad(date.month)}-${pad(date.day)}T${pad(h)}:${pad(m)}`;
	}

	function handleDateChange(newDate: DateValue | undefined) {
		selectedDate = newDate;
		if (newDate && (!value || (hour === '00' && minute === '00'))) {
			const now = new Date();
			const pad = (n: number) => n.toString().padStart(2, '0');
			hour = pad(now.getHours());
			minute = pad(now.getMinutes());
		}
		updateValue(newDate, hour, minute);
	}

	function handleHourChange(val: string) {
		const num = Math.max(0, Math.min(23, parseInt(val, 10) || 0));
		hour = num.toString().padStart(2, '0');
		updateValue(selectedDate, hour, minute);
	}

	function handleMinuteChange(val: string) {
		const num = Math.max(0, Math.min(59, parseInt(val, 10) || 0));
		minute = num.toString().padStart(2, '0');
		updateValue(selectedDate, hour, minute);
	}

	function setNow() {
		const now = new Date();
		const pad = (n: number) => n.toString().padStart(2, '0');
		selectedDate = new CalendarDate(now.getFullYear(), now.getMonth() + 1, now.getDate());
		hour = pad(now.getHours());
		minute = pad(now.getMinutes());
		updateValue(selectedDate, hour, minute);
	}

	let displayLabel = $derived.by(() => {
		if (!selectedDate) return '';
		const pad = (n: number) => n.toString().padStart(2, '0');
		const monthNames = [
			'Jan',
			'Feb',
			'Mar',
			'Apr',
			'May',
			'Jun',
			'Jul',
			'Aug',
			'Sep',
			'Oct',
			'Nov',
			'Dec'
		];
		const monthName = monthNames[selectedDate.month - 1] ?? pad(selectedDate.month);
		return `${pad(selectedDate.day)} ${monthName} ${selectedDate.year}, ${hour}:${minute}`;
	});
</script>

<Popover.Root bind:open={isOpen}>
	<Popover.Trigger
		{id}
		{disabled}
		class={cn(
			'w-full h-9 px-3 rounded-md border border-border bg-background/50 hover:bg-background/80 text-foreground text-xs font-normal transition-colors flex items-center justify-between gap-2 text-left outline-none focus:ring-1 focus:ring-primary focus:border-primary disabled:cursor-not-allowed disabled:opacity-50 select-none cursor-pointer',
			className
		)}
	>
		<div class="flex items-center gap-2 truncate">
			<CalendarIcon class="size-3.5 shrink-0 text-muted-foreground" />
			{#if displayLabel}
				<span class="truncate">{displayLabel}</span>
			{:else}
				<span class="text-muted-foreground truncate">{placeholder}</span>
			{/if}
		</div>
		<Clock class="size-3.5 shrink-0 text-muted-foreground" />
	</Popover.Trigger>

	<Popover.Content
		class="z-[60] w-auto p-3 bg-card border-border text-foreground space-y-3"
		align="start"
	>
		<Calendar
			type="single"
			bind:value={selectedDate as never}
			onValueChange={handleDateChange}
			class="rounded-md border-0 p-0"
		/>

		<div class="pt-3 border-t border-border flex items-center justify-between gap-3">
			<div class="flex items-center gap-1.5 text-xs text-muted-foreground font-medium">
				<Clock class="size-3.5 text-muted-foreground" />
				<span>Time</span>
			</div>

			<div class="flex items-center gap-1.5">
				<div
					class="flex items-center gap-1 bg-background border border-border rounded-md px-1.5 py-0.5"
				>
					<input
						type="number"
						min="0"
						max="23"
						value={hour}
						oninput={(e) => handleHourChange(e.currentTarget.value)}
						class="w-7 h-7 text-center text-xs font-mono bg-transparent text-foreground focus:outline-none"
						aria-label="Hour (00-23)"
					/>
					<span class="text-xs font-semibold text-muted-foreground">:</span>
					<input
						type="number"
						min="0"
						max="59"
						value={minute}
						oninput={(e) => handleMinuteChange(e.currentTarget.value)}
						class="w-7 h-7 text-center text-xs font-mono bg-transparent text-foreground focus:outline-none"
						aria-label="Minute (00-59)"
					/>
				</div>

				<button
					type="button"
					onclick={setNow}
					class="text-[11px] font-medium text-primary hover:text-primary/80 transition-colors px-1.5 py-1 rounded hover:bg-primary/10 cursor-pointer"
				>
					Now
				</button>
			</div>
		</div>
	</Popover.Content>
</Popover.Root>
