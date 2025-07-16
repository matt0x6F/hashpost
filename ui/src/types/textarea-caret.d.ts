declare module 'textarea-caret' {
  interface CaretCoordinates {
    left: number;
    top: number;
  }
  
  function getCaretCoordinates(
    element: HTMLTextAreaElement,
    position: number
  ): CaretCoordinates;
  
  export = getCaretCoordinates;
} 