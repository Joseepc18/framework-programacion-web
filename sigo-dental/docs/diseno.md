# Dirección visual: Aurora

Interfaz oscura con paneles de vidrio esmerilado, sobria y con un solo color de acento. La idea es que no parezca un panel genérico de tarjetas blancas: el odontograma es el elemento protagonista y los colores tienen significado clínico.

## Principios

- **Un solo acento.** El turquesa es el único color de marca. El coral se reserva para alertas y caries, y el ámbar para endodoncia.
- **Vidrio contenido.** Los paneles son casi opacos, con desenfoque suave. El efecto da profundidad, pero el texto siempre se lee primero.
- **El color significa algo.** Ningún color es decorativo en las pantallas clínicas.
- **Datos densos, superficie más sólida.** Tablas y listas largas usan fondo más opaco que las tarjetas.

## Fondo

Color base `#070B16` con dos manchas de luz desenfocadas, detrás del contenido:

| Mancha | Color | Opacidad | Tamaño | Desenfoque | Posición |
|---|---|---|---|---|---|
| Turquesa | `#14B8A6` | 0,30 | 540 px | 120 px | arriba a la izquierda |
| Violeta | `#7C3AED` | 0,28 | 580 px | 130 px | a la derecha |

## Superficies (vidrio)

- Panel: fondo `rgba(16, 22, 40, 0.72)`, borde `1px solid rgba(255, 255, 255, 0.10)`, radio 20 px.
- Desenfoque: `backdrop-filter: blur(10px) saturate(110%)`.
- Tablas y listas densas: fondo `rgba(16, 22, 40, 0.90)`.
- Alternativa obligatoria: si el navegador no soporta `backdrop-filter`, o el usuario pide menos transparencia, el panel usa un fondo sólido `#10162A` sin desenfoque.

## Color

| Uso | Valor |
|---|---|
| Texto principal | `#F5F7FF` |
| Texto secundario | `rgba(245, 247, 255, 0.72)` |
| Acento (botón principal, ítem activo, "En atención") | `#2DD4BF` |
| Texto sobre el acento | `#04201C` |
| Alerta crítica (chip) | fondo `rgba(251, 113, 133, 0.20)`, borde `rgba(251, 113, 133, 0.55)`, texto `#FFD1D8` |
| Valor de alerta numérico | `#FDA4AF` |

### Estados del odontograma

| Estado | Relleno | Borde |
|---|---|---|
| Sano | `rgba(255, 255, 255, 0.92)` | `rgba(255, 255, 255, 0.95)` |
| Caries | `#FB7185` | `#FDA4AF` |
| Restauración | `#60A5FA` (azul, como pide RF-03) | `#BFDBFE` |
| Sellante | `#2DD4BF` | `#5EEAD4` |
| Implante | `#A78BFA` | `#DDD6FE` |
| Prótesis fija / removible | `#C4B5FD` / `#F9A8D4` (marco) | — |
| Exodoncia | cruz `#FB7185` | — |
| Endodoncia | `#FBBF24` | `#FDE68A` |
| Ausente | transparente | punteado `rgba(255, 255, 255, 0.4)` |

El color nunca es el único indicador: cada hallazgo tiene leyenda y forma propia (triángulo, tornillo, cruz, marco, contorno punteado) y el estado se distingue por el trazo (punteado = requerido, grueso = realizado).

### Estados de la cita

| Estado | Relleno | Borde | Texto |
|---|---|---|---|
| Programado y confirmado | `rgba(255, 255, 255, 0.08)` | `rgba(255, 255, 255, 0.22)` | `rgba(245, 247, 255, 0.88)` |
| En atención | `rgba(45, 212, 191, 0.20)` | `rgba(45, 212, 191, 0.60)` | `#99F6E4` |
| Atendido | `rgba(255, 255, 255, 0.05)` | `rgba(255, 255, 255, 0.14)` | `rgba(245, 247, 255, 0.62)` |

Los estados Programado y Confirmado, antes agrupados, se distinguen en la agenda, y se añaden los tres que faltaban. Cada estado muestra siempre su nombre (el color solo lo refuerza); Cancelado además va tachado y con borde punteado.

| Estado | Relleno | Borde | Texto |
|---|---|---|---|
| Programado | `rgba(255, 255, 255, 0.08)` | `rgba(255, 255, 255, 0.22)` | `rgba(245, 247, 255, 0.88)` |
| Confirmado | `rgba(45, 212, 191, 0.08)` | `rgba(45, 212, 191, 0.45)` | `#CCFBF1` |
| En sala de espera | `rgba(255, 255, 255, 0.16)` | `rgba(255, 255, 255, 0.50)` | `#F5F7FF` |
| Cancelado | transparente | punteado `rgba(255, 255, 255, 0.28)` | `rgba(245, 247, 255, 0.50)`, tachado |
| No asistió | `rgba(251, 113, 133, 0.10)` | `rgba(251, 113, 133, 0.40)` | `#FFD1D8` |

## Tipografía

- **Sora** (400, 500 y 600) para la interfaz.
- **JetBrains Mono** (400 y 500) para horas, números de pieza dental (FDI) y número de historia clínica.
- Escala: título de página 30 px (600), título de sección 18 px (600), cuerpo 14 a 15 px, etiquetas 12 a 13 px.

## Forma

- Radio de tarjetas 20 px, controles 14 px y chips completamente redondeados.
- Barra lateral y tarjetas con el mismo vidrio; el ítem activo sube a `rgba(255, 255, 255, 0.14)` con borde `rgba(255, 255, 255, 0.24)`.

## Accesibilidad

- El texto sobre vidrio debe cumplir contraste 4,5:1 como mínimo.
- El color nunca es el único indicador de estado: cada estado lleva también texto o forma.
- Los controles táctiles miden 44 px de alto o más, por el uso en tablet (RNF-07).

## Aplicación en el frontend

Los valores se declaran como tokens en el bloque `@theme` de `frontend/src/index.css` (Tailwind 4) y los componentes los consumen; no se escriben colores sueltos en los componentes.
