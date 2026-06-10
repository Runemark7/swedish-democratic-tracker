export interface KpiMeta {
  label: string;
  description: string;
  unit: string;
  worseHigher: boolean;
  format: (v: number) => string;
}

export const STRIP_KPI_META: Record<string, KpiMeta> = {
  N00900: {
    label: "Kommunalskatt",
    description: "Din inkomstskatt till kommunen. Lägre skatt ger mer kvar i plånboken — men kan också innebära sämre service.",
    unit: "%", worseHigher: true, format: v => `${v.toFixed(2)} %`,
  },
  N03102: {
    label: "Resultat/skatt",
    description: "Kommunens överskott i förhållande till skatteintäkterna — ett mått på hur stor marginal kommunen har i sin ekonomi.",
    unit: "%", worseHigher: false, format: v => `${v.toFixed(1)} %`,
  },
  N03106: {
    label: "Soliditet",
    description: "Hur stor del av kommunens tillgångar som är skuldfria. Låg soliditet ökar sårbarheten vid ekonomiska kriser.",
    unit: "%", worseHigher: false, format: v => `${v.toFixed(0)} %`,
  },
  N15428: {
    label: "Gymnasiebehörighet",
    description: "Andel elever i åk 9 som är behöriga till gymnasiet. Viktig signal om skolkvaliteten i kommunen.",
    unit: "%", worseHigher: false, format: v => `${v.toFixed(1)} %`,
  },
  N00708: {
    label: "Arbetslöshet",
    description: "Andel av befolkningen 20–64 år som var arbetslösa någon gång under året. Låg arbetslöshet stärker kommunens skatteunderlag.",
    unit: "%", worseHigher: true, format: v => `${v.toFixed(1)} %`,
  },
};

export const STRIP_ORDER = ["N00900", "N03102", "N03106", "N15428", "N00708"];
