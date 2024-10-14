import {
  useCallback,
  useEffect,
  useMemo,
  useState,
  type PointerEvent,
} from "react";
import styles from "./Board.module.css";

type Props = {
  grid: Record<string, string>;
  offsetX?: number;
  offsetY?: number;
  width: number;
  height: number;
  scale?: number;
  fullScreen?: boolean;
  allowPanning?: boolean;
  cellSize?: number;
  padding?: number;
};

export default function Board({
  grid,
  offsetX: initialOffsetX = 0,
  offsetY: initialOffsetY = 0,
  width,
  height,
  scale = 1,
  fullScreen = false,
  allowPanning = false,
  cellSize = 50,
  padding = 10,
}: Props) {
  const [startX, setStartX] = useState<number | null>(null);
  const [startY, setStartY] = useState<number | null>(null);
  const [offsetX, setOffsetX] = useState(initialOffsetX);
  const [displacementX, setDisplacementX] = useState(0);
  const [offsetY, setOffsetY] = useState(initialOffsetY);
  const [displacementY, setDisplacementY] = useState(0);

  const allPointsMemoized = useMemo(() => {
    return allPoints(grid, offsetX, offsetY, width, height, cellSize);
  }, [grid, offsetX, offsetY, width, height, cellSize]);

  const startDrag = useCallback((e: PointerEvent) => {
    setStartX(e.clientX);
    setStartY(e.clientY);
  }, []);

  const drag = useCallback(
    (e: PointerEvent) => {
      if (startX !== null && startY !== null && allowPanning) {
        setDisplacementX(startX - e.clientX);
        setDisplacementY(startY - e.clientY);
      }
    },
    [startX, startY, allowPanning]
  );

  const stopDrag = useCallback(
    (e: PointerEvent) => {
      setOffsetX((prevOffsetX) => prevOffsetX + displacementX);
      setOffsetY((prevOffsetY) => prevOffsetY + displacementY);
      setDisplacementX(0);
      setDisplacementY(0);
      setStartX(null);
      setStartY(null);
    },
    [displacementX, displacementY]
  );

  useEffect(() => {
    const doNothing = () => {};
    window.addEventListener("pointermove", doNothing);
    return () => {
      window.removeEventListener("pointermove", doNothing);
    };
  });

  return (
    <svg
      className={styles.board}
      style={{
        touchAction: "none",
        ...(fullScreen
          ? {
              position: "fixed",
              top: 0,
              left: 0,
              width: "100%",
              height: "100%",
            }
          : {
              width: width + padding,
              height: height + padding,
            }),
      }}
      viewBox={`${(offsetX + displacementX) / scale - padding / 2} ${(offsetY + displacementY) / scale - padding / 2} ${width / scale + padding} ${height / scale + padding}`}
      onPointerDown={startDrag}
      onPointerMove={drag}
      onPointerUp={stopDrag}
    >
      {allPointsMemoized.map(([x, y]) => {
        const key = `${x},${y}`;
        const value = grid[key];
        const isModifier =
          value === "DW" || value === "DL" || value === "TW" || value === "TL";

        return (
          <g
            key={key}
            fontSize={cellSize / 2}
            className={styles.cell}
            style={{
              transformOrigin: `${(x + 0.5) * cellSize}px ${(y + 0.5) * cellSize}px`,
              animationDelay: `${Math.random() * 200}ms`,
            }}
          >
            <rect
              x={x * cellSize + 3}
              y={y * cellSize + 3}
              width={cellSize - 6}
              height={cellSize - 6}
              stroke="#444"
              rx="7"
              ry="7"
              fill={boxFill(value)}
            />
            {isModifier ? (
              <text
                x={(x + 0.5) * cellSize}
                y={(y + 0.5) * cellSize}
                textAnchor="middle"
                dominantBaseline="central"
                fontSize={cellSize / 3}
              >
                {value}
              </text>
            ) : (
              <text
                x={(x + 0.5) * cellSize}
                y={(y + 0.5) * cellSize}
                textAnchor="middle"
                dominantBaseline="central"
              >
                {value}
              </text>
            )}
          </g>
        );
      })}
    </svg>
  );
}

function boxFill(value: string): string {
  if (!value) {
    return "rgba(255,255,255,0.25)";
  }

  switch (value) {
    case "DW":
      return "#faf";
    case "DL":
      return "#aaf";
    case "TW":
      return "#faa";
    case "TL":
      return "#faa";
  }

  return "rgba(255,255,255,0.9)";
}

type point = [number, number];

function allPoints(
  grid: Record<string, string>,
  offsetX: number,
  offsetY: number,
  width: number,
  height: number,
  cellSize: number
): point[] {
  let minX = Infinity;
  let maxX = -Infinity;
  let minY = Infinity;
  let maxY = -Infinity;

  for (const point of Object.keys(grid)) {
    const [x, y] = point.split(",").map(Number);
    if (x < minX) minX = x;
    if (x > maxX) maxX = x;
    if (y < minY) minY = y;
    if (y > maxY) maxY = y;
  }

  const points: point[] = [];
  for (let x = minX; x <= maxX; x++) {
    for (let y = minY; y <= maxY; y++) {
      // only push points that should be visible
      if (
        x * cellSize - offsetX >= -cellSize * 5 &&
        x * cellSize - offsetX <= width + 4 * cellSize &&
        y * cellSize - offsetY >= -cellSize * 5 &&
        y * cellSize - offsetY <= height + 4 * cellSize
      ) {
        points.push([x, y]);
      }
    }
  }

  return points;
}

function isModifier(value: string): boolean {
  return value === "DW" || value === "DL" || value === "TW" || value === "TL";
}
