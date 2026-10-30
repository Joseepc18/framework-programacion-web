import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { Aviso, BotonEnlace, Campo, Chip, Tabla, Td, Th, Titulo } from '../components/ui';
import { useAuth } from '../context/useAuth';
import { useCarga } from '../hooks/useCarga';
import { ETIQUETA_ESTADO_PAGO, listarCobros } from '../services/cobroService';
import { formatearDinero } from '../utils/dinero';

const formatoFecha = new Intl.DateTimeFormat('es', { day: '2-digit', month: 'short', year: 'numeric' });

const CobrosPage = () => {
  const { usuario } = useAuth();
  const [texto, setTexto] = useState('');
  const [busqueda, setBusqueda] = useState('');

  // La búsqueda se envía 300 ms después de dejar de escribir.
  useEffect(() => {
    const espera = setTimeout(() => setBusqueda(texto), 300);
    return () => clearTimeout(espera);
  }, [texto]);

  const { datos, error, cargando } = useCarga(() => listarCobros(busqueda), `cobros|${busqueda}`);

  return (
    <section aria-labelledby="titulo-cobros" className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <Titulo id="titulo-cobros">Cobros</Titulo>
        {usuario?.rol === 'recepcionista' && <BotonEnlace to="/cobros/nuevo">Nuevo cobro</BotonEnlace>}
      </div>

      <div className="max-w-xl">
        <Campo id="buscar-cobro" type="search" etiqueta="Buscar por comprobante, paciente o documento" value={texto} onChange={(e) => setTexto(e.target.value)} />
      </div>

      {error && <Aviso>{error}</Aviso>}
      {cargando && !datos && <p>Cargando…</p>}
      {datos && datos.length === 0 && <Aviso tipo="info">No hay cobros que coincidan con la búsqueda.</Aviso>}
      {datos && datos.length > 0 && (
        <Tabla etiqueta="Cobros">
          <thead>
            <tr><Th>Comprobante</Th><Th>Fecha</Th><Th>Paciente</Th><Th>Total</Th><Th>Pagado</Th><Th>Saldo</Th><Th>Estado</Th><Th>Abrir</Th></tr>
          </thead>
          <tbody>
            {datos.map((c) => (
              <tr key={c.id} className={c.anulado ? 'opacity-60' : ''}>
                <Td className="font-mono">{c.numero}</Td>
                <Td>{formatoFecha.format(new Date(c.creado_en))}</Td>
                <Td>{c.paciente_nombre}</Td>
                <Td className="font-mono">{formatearDinero(c.total_centavos)}</Td>
                <Td className="font-mono">{formatearDinero(c.pagado_centavos)}</Td>
                <Td className="font-mono">{formatearDinero(c.saldo_centavos)}</Td>
                <Td>
                  {c.anulado ? <Chip tono="alerta">Anulado</Chip> : <Chip tono={c.estado === 'pagado' ? 'acento' : 'neutro'}>{ETIQUETA_ESTADO_PAGO[c.estado]}</Chip>}
                </Td>
                <Td><Link className="text-chip-acento-texto underline" to={`/cobros/${c.id}`}>Abrir<span className="sr-only"> {c.numero}</span></Link></Td>
              </tr>
            ))}
          </tbody>
        </Tabla>
      )}
    </section>
  );
};

export default CobrosPage;
