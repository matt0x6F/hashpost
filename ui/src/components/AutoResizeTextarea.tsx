import React, { useRef, useEffect } from 'react';
import { Textarea } from './shadcn/textarea';

interface AutoResizeTextareaProps extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
  value: string;
  onChange: (e: React.ChangeEvent<HTMLTextAreaElement>) => void;
  minHeight?: number;
  maxHeight?: number;
}

export function AutoResizeTextarea({ 
  value, 
  onChange, 
  minHeight = 100, 
  maxHeight = 600,
  className = "",
  ...props 
}: AutoResizeTextareaProps) {
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const adjustHeight = () => {
    const textarea = textareaRef.current;
    if (!textarea) return;

    // Reset height to auto to get the correct scrollHeight
    textarea.style.height = 'auto';
    
    // Calculate the new height
    const scrollHeight = textarea.scrollHeight;
    const newHeight = Math.max(minHeight, Math.min(scrollHeight, maxHeight));
    
    textarea.style.height = `${newHeight}px`;
  };

  useEffect(() => {
    adjustHeight();
  }, [value, minHeight, maxHeight]);

  return (
    <Textarea
      ref={textareaRef}
      value={value}
      onChange={onChange}
      className={className}
      style={{ 
        resize: 'none',
        overflowY: 'auto',
        minHeight: `${minHeight}px`,
        maxHeight: `${maxHeight}px`
      }}
      {...props}
    />
  );
} 