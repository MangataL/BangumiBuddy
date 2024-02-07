import { CSSTransition } from "react-transition-group";
import { useRef } from "react";

export const PAGE_TRANSITION_DURATION = 300;

interface PageTransitionProps {
  children: React.ReactNode;
  fadeOut?: boolean;
}

export function PageTransition({
  children,
  fadeOut = false,
}: PageTransitionProps) {
  const nodeRef = useRef(null);

  return (
    <div className="w-full h-full">
      <CSSTransition
        nodeRef={nodeRef}
        in={!fadeOut}
        timeout={PAGE_TRANSITION_DURATION}
        classNames="page"
        unmountOnExit
      >
        <div ref={nodeRef} className="page-wrapper">
          {children}
        </div>
      </CSSTransition>
    </div>
  );
}
