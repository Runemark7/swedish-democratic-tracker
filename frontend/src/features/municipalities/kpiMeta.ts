export interface KpiMeta {
  label: string;
  description: string;
  unit: string;
  target: number;
  worseHigher: boolean;
  format: (v: number) => string;
}

export const STRIP_KPI_META: Record<string, KpiMeta> = {
  N00900: {
    label: "Kommunalskatt",
    description: "Din inkomstskatt till kommunen. Lägre skatt ger mer kvar i plånboken — men kan också innebära sämre service.",
    unit: "%", target: 31.0, worseHigher: true, format: v => `${v.toFixed(2)} %`,
  },
  N03102: {
    label: "Resultat/skatt",
    description: "Kommunens överskott i förhållande till skatteintäkterna. Under 2 % riskerar kommunen att tvingas skära i välfärden.",
    unit: "%", target: 2.0, worseHigher: false, format: v => `${v.toFixed(1)} %`,
  },
  N03106: {
    label: "Soliditet",
    description: "Hur stor del av kommunens tillgångar som är skuldfria. Låg soliditet ökar sårbarheten vid ekonomiska kriser.",
    unit: "%", target: 25.0, worseHigher: false, format: v => `${v.toFixed(0)} %`,
  },
  N15428: {
    label: "Gymnasiebehörighet",
    description: "Andel elever i åk 9 som är behöriga till gymnasiet. Viktig signal om skolkvaliteten i kommunen.",
    unit: "%", target: 85.0, worseHigher: false, format: v => `${v.toFixed(1)} %`,
  },
  N00708: {
    label: "Arbetslöshet",
    description: "Andel av befolkningen 20–64 år som var arbetslösa någon gång under året. Låg arbetslöshet stärker kommunens skatteunderlag.",
    unit: "%", target: 5.0, worseHigher: true, format: v => `${v.toFixed(1)} %`,
  },
};

export const STRIP_ORDER = ["N00900", "N03102", "N03106", "N15428", "N00708"];
