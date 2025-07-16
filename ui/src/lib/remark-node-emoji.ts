import { visit } from 'unist-util-visit';
import { Root, Text } from 'mdast';
import { getEmojiByShortcode } from './emojiRegistry';

export default function remarkNodeEmoji() {
  return (tree: Root) => {
    visit(tree, 'text', (node: Text) => {
      node.value = node.value.replace(/:([a-zA-Z0-9_+-]+):/g, (match: string, shortcode: string) => {
        return getEmojiByShortcode(shortcode) || match;
      });
    });
  };
} 