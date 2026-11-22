import { motion, useReducedMotion } from "motion/react";
import type { ReactNode } from "react";

export function PanelContent({
  children,
  visible = true,
}: {
  children: ReactNode;
  visible?: boolean;
}) {
  const reduced = useReducedMotion();
  return (
    <motion.div
      initial={false}
      animate={{ opacity: visible ? 1 : 0 }}
      transition={{ duration: reduced ? 0 : 0.14 }}
      className="flex flex-col gap-2 py-4.5 px-3.5 wrap-anywhere"
    >
      {children}
    </motion.div>
  );
}
