import { useRef, useState } from 'react';
import type { FirmaManuscrita, Punto } from '../services/consentimientoService';
import { Boton } from './ui';

const ANCHO = 600;
const ALTO = 220;

// Lienzo para la firma manuscrita con el dedo, el lápiz o el ratón. Guarda cada punto con su marca de tiempo
// (milisegundos desde el inicio del trazo). `touch-action: none` evita que la pantalla se desplace mientras se firma.
const LienzoFirma = ({ alConfirmar, ocupado }: { alConfirmar: (firma: FirmaManuscrita) => void; ocupado: boolean }) => {
  const lienzo = useRef<HTMLCanvasElement>(null);
  const trazos = useRef<Punto[][]>([]);
  const dibujando = useRef(false);
  const inicio = useRef(0);
  const [hayFirma, setHayFirma] = useState(false);

  // Convierte la posición del puntero al sistema de coordenadas interno del lienzo (600 x 220).
  const posicion = (e: React.PointerEvent<HTMLCanvasElement>): Punto => {
    const caja = e.currentTarget.getBoundingClientRect();
    return {
      x: Math.round(((e.clientX - caja.left) / caja.width) * ANCHO * 10) / 10,
      y: Math.round(((e.clientY - caja.top) / caja.height) * ALTO * 10) / 10,
      t: Math.round(performance.now() - inicio.current),
    };
  };

  const empezar = (e: React.PointerEvent<HTMLCanvasElement>) => {
    e.currentTarget.setPointerCapture(e.pointerId);
    if (trazos.current.length === 0) inicio.current = performance.now();
    dibujando.current = true;
    const p = posicion(e);
    trazos.current.push([p]);
    const ctx = lienzo.current?.getContext('2d');
    if (ctx) {
      ctx.beginPath();
      ctx.arc(p.x, p.y, 1.1, 0, Math.PI * 2);
      ctx.fill();
    }
  };

  const mover = (e: React.PointerEvent<HTMLCanvasElement>) => {
    if (!dibujando.current) return;
    const actual = trazos.current[trazos.current.length - 1];
    const previo = actual[actual.length - 1];
    const p = posicion(e);
    actual.push(p);
    const ctx = lienzo.current?.getContext('2d');
    if (ctx) {
      ctx.beginPath();
      ctx.moveTo(previo.x, previo.y);
      ctx.lineTo(p.x, p.y);
      ctx.stroke();
    }
    setHayFirma(true);
  };

  const terminar = () => {
    dibujando.current = false;
  };

  const borrar = () => {
    trazos.current = [];
    setHayFirma(false);
    lienzo.current?.getContext('2d')?.clearRect(0, 0, ANCHO, ALTO);
  };

  // Al montar el canvas se fija el estilo del trazo (negro sobre fondo blanco, legible en el PDF).
  const preparar = (canvas: HTMLCanvasElement | null) => {
    lienzo.current = canvas;
    const ctx = canvas?.getContext('2d');
    if (ctx) {
      ctx.lineWidth = 2.4;
      ctx.lineCap = 'round';
      ctx.lineJoin = 'round';
      ctx.strokeStyle = '#111';
      ctx.fillStyle = '#111';
    }
  };

  return (
    <div className="space-y-3">
      <p className="text-etiqueta text-texto-suave">Firma con el dedo o el lápiz dentro del recuadro.</p>
      <canvas
        ref={preparar}
        width={ANCHO}
        height={ALTO}
        role="img"
        aria-label="Recuadro para dibujar la firma"
        onPointerDown={empezar}
        onPointerMove={mover}
        onPointerUp={terminar}
        onPointerCancel={terminar}
        className="aspect-[600/220] w-full max-w-xl touch-none rounded-control border border-chip-neutro-borde bg-white"
      />
      <div className="flex flex-wrap gap-3">
        <Boton disabled={!hayFirma || ocupado} onClick={() => alConfirmar({ ancho: ANCHO, alto: ALTO, trazos: trazos.current })}>
          {ocupado ? 'Guardando…' : 'Confirmar firma'}
        </Boton>
        <Boton variante="secundario" disabled={ocupado} onClick={borrar}>Borrar</Boton>
      </div>
    </div>
  );
};

export default LienzoFirma;
