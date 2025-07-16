import emojiList from './emojiList.json';

// Add custom emojis and aliases here if needed
const customEmojis = [
  { shortcode: 'hashpost', emoji: '🛡️', name: 'HashPost' },
  // ...add more custom emojis or aliases
];

export function getAllEmojis() {
  // Avoid duplicate shortcodes
  const customShortcodes = new Set(customEmojis.map(e => e.shortcode));
  const filteredBase = (emojiList as Array<{shortcode: string, emoji: string, name: string}>).filter(e => !customShortcodes.has(e.shortcode));
  return [...filteredBase, ...customEmojis];
}

export function getEmojiByShortcode(shortcode: string) {
  const found = getAllEmojis().find(e => e.shortcode === shortcode);
  return found ? found.emoji : undefined;
} 