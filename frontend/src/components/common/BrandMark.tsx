interface BrandMarkProps {
  size?: number;
}

// A routed crossbar turns the A monogram into a small automation path.
export function BrandMark({ size = 30 }: BrandMarkProps) {
  return <img src="/automata-mark.svg" width={size} height={size} alt="" aria-hidden="true" draggable={false} />;
}
