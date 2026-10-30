import { useState } from 'react';
import { Titulo } from '../components/ui';
import PanelParametros from './catalogos/PanelParametros';
import PanelProcedimientos from './catalogos/PanelProcedimientos';
import PanelProfesionales from './catalogos/PanelProfesionales';
import PanelSillones from './catalogos/PanelSillones';
import PanelSimple from './catalogos/PanelSimple';

const PESTANAS = [
  { id: 'consultorios', etiqueta: 'Consultorios', panel: () => <PanelSimple recurso="consultorios" singular="consultorio" /> },
  { id: 'sillones', etiqueta: 'Sillones', panel: () => <PanelSillones /> },
  { id: 'profesionales', etiqueta: 'Profesionales', panel: () => <PanelProfesionales /> },
  { id: 'procedimientos', etiqueta: 'Procedimientos', panel: () => <PanelProcedimientos /> },
  { id: 'convenios', etiqueta: 'Convenios', panel: () => <PanelSimple recurso="convenios" singular="convenio" /> },
  { id: 'condiciones', etiqueta: 'Condiciones críticas', panel: () => <PanelSimple recurso="condiciones-criticas" singular="condición crítica" /> },
  { id: 'parametros', etiqueta: 'Parámetros', panel: () => <PanelParametros /> },
] as const;

// Parametrización de catálogos del rol Administrador TI. Nada se borra: se desactiva.
const CatalogosPage = () => {
  const [activa, setActiva] = useState<(typeof PESTANAS)[number]['id']>('consultorios');
  const pestana = PESTANAS.find((p) => p.id === activa) ?? PESTANAS[0];

  return (
    <section aria-labelledby="titulo-catalogos" className="space-y-6">
      <Titulo id="titulo-catalogos">Catálogos</Titulo>
      <div role="tablist" aria-label="Catálogos" className="flex flex-wrap gap-2">
        {PESTANAS.map((p) => (
          <button
            key={p.id}
            id={`tab-${p.id}`}
            role="tab"
            type="button"
            aria-selected={p.id === activa}
            aria-controls="panel-catalogo"
            onClick={() => setActiva(p.id)}
            className={`min-h-11 rounded-control border px-4 text-cuerpo transition focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-acento ${
              p.id === activa
                ? 'border-item-activo-borde bg-item-activo font-semibold text-texto'
                : 'border-transparent text-texto-suave hover:bg-chip-neutro-fondo hover:text-texto'
            }`}
          >
            {p.etiqueta}
          </button>
        ))}
      </div>
      <div id="panel-catalogo" role="tabpanel" aria-labelledby={`tab-${pestana.id}`}>
        {pestana.panel()}
      </div>
    </section>
  );
};

export default CatalogosPage;
