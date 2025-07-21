'use client';

import React, { useRef, useEffect, useState, useCallback } from 'react';
import { Textarea } from './shadcn/textarea';
import { Button } from './shadcn/button';
import { Smile } from 'lucide-react';
import { getAllEmojis } from '@/lib/emojiRegistry';
import getCaretCoordinates from 'textarea-caret';

interface MarkdownTextareaProps extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
  value: string;
  onChange: (e: React.ChangeEvent<HTMLTextAreaElement>) => void;
  minHeight?: number;
  maxHeight?: number;
  showEmojiPicker?: boolean;
  className?: string;
}

// Build emoji list from node-emoji using .search('')
const emojiList = getAllEmojis();

// Optionally, group by first letter or category if desired (for now, flat list)

type Emoji = { shortcode: string; emoji: string; name: string };

function EmojiPopover({ filteredEmojis, emojiSearch, insertEmojiShortcode }: {
  filteredEmojis: Emoji[];
  emojiSearch: string;
  insertEmojiShortcode: (shortcode: string) => void;
}) {
  const [hovered, setHovered] = React.useState<string | null>(null);
  const [visibleCount, setVisibleCount] = React.useState(32);
  const gridRef = React.useRef<HTMLDivElement>(null);

  // Reset visibleCount when search changes
  React.useEffect(() => {
    setVisibleCount(32);
  }, [emojiSearch, filteredEmojis]);

  // Lazy load more emojis on scroll
  const handleScroll = () => {
    const grid = gridRef.current;
    if (!grid) return;
    
    // Check if we're near the bottom (within 100px)
    const isNearBottom = grid.scrollTop + grid.clientHeight >= grid.scrollHeight - 100;
    

    
    if (isNearBottom && visibleCount < filteredEmojis.length) {
      // Load more emojis
      setVisibleCount((c) => {
        const next = Math.min(filteredEmojis.length, c + 32);
        return next;
      });
    }
  };

  const visibleEmojis = filteredEmojis.slice(0, visibleCount);
  const canLoadMore = visibleCount < filteredEmojis.length;

  return (
    <div
      className="w-full bg-background rounded-lg shadow-lg border border-border"
      style={{ minWidth: 200 }}
      role="listbox"
      aria-label="Emoji suggestions"
    >
      <div
        className="grid grid-cols-8 gap-1 max-h-32 overflow-y-auto"
        ref={gridRef}
        onScroll={handleScroll}
      >
        {visibleEmojis.map((emoji) => (
          <button
            key={emoji.shortcode}
            type="button"
            className="w-8 h-8 flex items-center justify-center hover:bg-muted rounded text-lg"
            onClick={() => insertEmojiShortcode(emoji.shortcode)}
            title={`:${emoji.shortcode}: ${emoji.name}`}
            tabIndex={-1}
            onMouseEnter={() => setHovered(emoji.shortcode)}
            onMouseLeave={() => setHovered(null)}
          >
            {emoji.emoji}
          </button>
        ))}
      </div>
      {/* Load More button outside scrollable area */}
      {canLoadMore && (
        <div className="p-2 border-t border-border">
          <button
            className="w-full py-2 bg-muted hover:bg-muted/80 rounded text-xs text-muted-foreground hover:text-foreground transition-colors"
            onClick={() => setVisibleCount((c) => Math.min(filteredEmojis.length, c + 32))}
            type="button"
          >
            Load more ({visibleCount} of {filteredEmojis.length})
          </button>
        </div>
      )}
      {filteredEmojis.length === 0 && (
        <div className="text-center text-muted-foreground py-4">
          No emojis found for &quot;{emojiSearch}&quot;
        </div>
      )}
      {/* Shortcode bar at the bottom */}
      <div className="w-full bg-background text-xs text-center py-1 border-t border-border rounded-b-lg shadow-lg">
        {hovered ? `:${hovered}:` : '\u00A0'}
      </div>
    </div>
  );
}

export function MarkdownTextarea({ 
  value, 
  onChange, 
  minHeight = 100, 
  maxHeight = 600,
  className = "",
  showEmojiPicker = true,
  ...props 
}: MarkdownTextareaProps) {
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const [showEmojiPopover, setShowEmojiPopover] = useState(false);
  const [emojiSearch, setEmojiSearch] = useState('');
  const [currentColonIndex, setCurrentColonIndex] = useState<number | null>(null);
  const [pickerPos, setPickerPos] = useState<{left: number, top: number} | null>(null);
  const justClosedPopoverRef = useRef(false);

  const PICKER_WIDTH = 320;

  const adjustHeight = useCallback(() => {
    const textarea = textareaRef.current;
    if (!textarea) return;
    textarea.style.height = 'auto';
    const scrollHeight = textarea.scrollHeight;
    const newHeight = Math.max(minHeight, Math.min(scrollHeight, maxHeight));
    textarea.style.height = `${newHeight}px`;
  }, [minHeight, maxHeight]);

  useEffect(() => {
    adjustHeight();
  }, [value, adjustHeight]);

  // Handle emoji shortcode insertion
  const insertEmojiShortcode = useCallback((shortcode: string) => {
    if (!textareaRef.current || currentColonIndex === null) return;
    const textarea = textareaRef.current;
    const end = textarea.selectionEnd;
    // Replace from colon to cursor with :shortcode:
    const beforeColon = value.substring(0, currentColonIndex);
    const afterSelection = value.substring(end);
    const newValue = beforeColon + `:${shortcode}:` + afterSelection;
    const newCursor = beforeColon.length + shortcode.length + 2; // +2 for the colons
    const event = { target: { value: newValue } } as React.ChangeEvent<HTMLTextAreaElement>;
    onChange(event);
    setTimeout(() => {
      if (textareaRef.current) {
        textareaRef.current.setSelectionRange(newCursor, newCursor);
        textareaRef.current.focus();
      }
    }, 0);
    setShowEmojiPopover(false);
    setCurrentColonIndex(null);
    setEmojiSearch('');
  }, [value, onChange, currentColonIndex]);

  // When opening the emoji picker, calculate caret position
  const openEmojiPopoverAtCaret = useCallback(() => {
    if (textareaRef.current) {
      const textarea = textareaRef.current;
      const coords = getCaretCoordinates(textarea, textarea.selectionStart);
      setPickerPos({
        left: coords.left - textarea.scrollLeft - PICKER_WIDTH / 2,
        top: coords.top - textarea.scrollTop + 24,
      });
    } else {
      setPickerPos(null);
    }
    setShowEmojiPopover(true);
  }, []);

  // Update picker position on input
  useEffect(() => {
    if (showEmojiPopover && textareaRef.current) {
      const textarea = textareaRef.current;
      const coords = getCaretCoordinates(textarea, textarea.selectionStart);
      setPickerPos({
        left: coords.left - textarea.scrollLeft - PICKER_WIDTH / 2,
        top: coords.top - textarea.scrollTop + 24,
      });
    }
  }, [showEmojiPopover, value]);

  // Only open emoji picker when : is typed
  const handleKeyDown = useCallback((e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === ':') {
      setTimeout(() => {
        const textarea = textareaRef.current;
        if (!textarea) return;
        const textBeforeCursor = textarea.value.substring(0, textarea.selectionStart);
        if (/:[a-zA-Z0-9_]+:$/.test(textBeforeCursor)) {
          // Just completed a shortcode, do not open popover
          return;
        }
        setCurrentColonIndex(textarea.selectionStart - 1); // position of the colon
        setEmojiSearch('');
        openEmojiPopoverAtCaret();
      }, 0);
    } else if (e.key === 'Enter' || e.key === 'Escape') {
      setShowEmojiPopover(false);
      setCurrentColonIndex(null);
      setEmojiSearch('');
    } else if ((e.key === ' ' || e.key === 'Tab') && showEmojiPopover) {
      setShowEmojiPopover(false);
      setCurrentColonIndex(null);
      setEmojiSearch('');
    }
  }, [showEmojiPopover, openEmojiPopoverAtCaret]);

  // When typing after colon, update emoji search
  const handleTextareaChange = useCallback((e: React.ChangeEvent<HTMLTextAreaElement>) => {
    onChange(e);
    if (showEmojiPopover && currentColonIndex !== null) {
      const cursorPosition = e.target.selectionStart;
      let afterColon = e.target.value.substring(currentColonIndex + 1, cursorPosition);
      if (afterColon.startsWith(':')) {
        afterColon = afterColon.slice(1);
      }
      // If user types whitespace, a new line, or a second colon (after at least one character), close picker and reset all state
      if (/\s/.test(afterColon) || afterColon.indexOf(':') > 0) {
        setShowEmojiPopover(false);
        setCurrentColonIndex(null);
        setEmojiSearch('');
        justClosedPopoverRef.current = true;
        return;
      }
      setEmojiSearch(afterColon);
    }
  }, [onChange, showEmojiPopover, currentColonIndex]);

  // When the smiley face button is clicked, open the emoji picker with empty search
  const handleSmileyClick = useCallback(() => {
    if (textareaRef.current) {
      const cursorPosition = textareaRef.current.selectionStart;
      setCurrentColonIndex(cursorPosition);
      setEmojiSearch('');
      openEmojiPopoverAtCaret();
      textareaRef.current.focus();
    }
  }, [openEmojiPopoverAtCaret]);

  // Filter emojis based on search
  const filteredEmojis = emojiList.filter(emoji =>
    emoji.shortcode.toLowerCase().includes(emojiSearch.toLowerCase()) ||
    emoji.name.toLowerCase().includes(emojiSearch.toLowerCase())
  );

  return (
    <div className="relative">
      <Textarea
        ref={textareaRef}
        value={value}
        onChange={handleTextareaChange}
        onKeyDown={handleKeyDown}
        className={className}
        style={{ 
          resize: 'none',
          overflowY: 'auto',
          minHeight: `${minHeight}px`,
          maxHeight: `${maxHeight}px`
        }}
        {...props}
      />
      {/* Always show the smiley face button */}
      <div className="absolute bottom-2 right-2 z-50">
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className="h-8 w-8 p-0"
          aria-label="Open emoji picker"
          onClick={handleSmileyClick}
        >
          <Smile className="h-4 w-4" />
        </Button>
      </div>
      {/* Custom EmojiPopover at caret position */}
      {showEmojiPicker && showEmojiPopover && (
        <div
          style={{
            position: 'absolute',
            left: pickerPos ? pickerPos.left : 'auto',
            top: pickerPos ? pickerPos.top : '100%',
            right: pickerPos ? 'auto' : 0,
            zIndex: 100,
            width: PICKER_WIDTH,
          }}
        >
          <EmojiPopover
            filteredEmojis={filteredEmojis}
            emojiSearch={emojiSearch}
            insertEmojiShortcode={insertEmojiShortcode}
          />
        </div>
      )}
    </div>
  );
} 