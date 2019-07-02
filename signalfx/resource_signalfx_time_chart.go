package signalfx

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"

	chart "github.com/signalfx/signalfx-go/chart"
)

var PaletteColors = map[string]int{
	"gray":       0,
	"blue":       1,
	"azure":      2,
	"navy":       3,
	"brown":      4,
	"orange":     5,
	"yellow":     6,
	"magenta":    7,
	"purple":     8,
	"pink":       9,
	"violet":     10,
	"lilac":      11,
	"iris":       12,
	"emerald":    13,
	"green":      14,
	"aquamarine": 15,
}

var FullPaletteColors = map[string]int{
	"gray":        0,
	"blue":        1,
	"azure":       2,
	"navy":        3,
	"brown":       4,
	"orange":      5,
	"yellow":      6,
	"magenta":     7,
	"purple":      8,
	"pink":        9,
	"violet":      10,
	"lilac":       11,
	"iris":        12,
	"emerald":     13,
	"green":       14,
	"aquamarine":  15,
	"red":         16,
	"gold":        17,
	"greenyellow": 18,
	"chartreuse":  19,
	"jade":        20,
}

func resourceAxisMigrateState(v int, is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
	switch v {
	case 0:
		return migrateAxisStateV0toV1(is)
	default:
		return is, fmt.Errorf("Unexpected schema version: %d", v)
	}
}

func migrateAxisStateV0toV1(is *terraform.InstanceState) (*terraform.InstanceState, error) {
	if is.Empty() || is.Attributes == nil {
		return is, nil
	}
	if v, ok := is.Attributes["max_value"]; ok {
		if f, err := strconv.ParseFloat(v, 32); err == nil && f == math.MaxFloat32 {
			delete(is.Attributes, "max_value")
		}
	}
	if v, ok := is.Attributes["min_value"]; ok {
		if f, err := strconv.ParseFloat(v, 32); err == nil && f == -math.MaxFloat32 {
			delete(is.Attributes, "min_value")
		}
	}
	if v, ok := is.Attributes["low_watermark"]; ok {
		if f, err := strconv.ParseFloat(v, 32); err == nil && f == -math.MaxFloat32 {
			delete(is.Attributes, "low_watermark")
		}
	}
	if v, ok := is.Attributes["high_watermark"]; ok {
		if f, err := strconv.ParseFloat(v, 32); err == nil && f == math.MaxFloat32 {
			delete(is.Attributes, "high_watermark")
		}
	}
	return is, nil
}

func timeChartResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the chart",
			},
			"program_text": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Signalflow program text for the chart. More info at \"https://developers.signalfx.com/docs/signalflow-overview\"",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the chart",
			},
			"unit_prefix": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "Metric",
				Description: "(Metric by default) Must be \"Metric\" or \"Binary\"",
			},
			"color_by": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "Dimension",
				Description: "(Dimension by default) Must be \"Dimension\" or \"Metric\"",
			},
			"minimum_resolution": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The minimum resolution (in seconds) to use for computing the underlying program",
			},
			"max_delay": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "How long (in seconds) to wait for late datapoints",
				ValidateFunc: validateMaxDelayValue,
			},
			"disable_sampling": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "(false by default) If false, samples a subset of the output MTS, which improves UI performance",
			},
			"time_range": &schema.Schema{
				Type:          schema.TypeInt,
				Optional:      true,
				Description:   "Seconds to display in the visualization. This is a rolling range from the current time. Example: 8600 = `-1h`",
				ConflictsWith: []string{"start_time", "end_time"},
			},
			"start_time": &schema.Schema{
				Type:          schema.TypeInt,
				Optional:      true,
				Description:   "Seconds since epoch to start the visualization",
				ConflictsWith: []string{"time_range"},
			},
			"end_time": &schema.Schema{
				Type:          schema.TypeInt,
				Optional:      true,
				Description:   "Seconds since epoch to end the visualization",
				ConflictsWith: []string{"time_range"},
			},
			"axis_right": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					SchemaVersion: 1,
					MigrateState:  resourceAxisMigrateState,
					Schema: map[string]*schema.Schema{
						"min_value": &schema.Schema{
							Type:        schema.TypeFloat,
							Optional:    true,
							Default:     -math.MaxFloat64,
							Description: "The minimum value for the right axis",
						},
						"max_value": &schema.Schema{
							Type:        schema.TypeFloat,
							Optional:    true,
							Default:     math.MaxFloat64,
							Description: "The maximum value for the right axis",
						},
						"label": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Label of the right axis",
						},
						"high_watermark": &schema.Schema{
							Type:        schema.TypeFloat,
							Optional:    true,
							Default:     math.MaxFloat64,
							Description: "A line to draw as a high watermark",
						},
						"high_watermark_label": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "A label to attach to the high watermark line",
						},
						"low_watermark": &schema.Schema{
							Type:        schema.TypeFloat,
							Optional:    true,
							Default:     -math.MaxFloat64,
							Description: "A line to draw as a low watermark",
						},
						"low_watermark_label": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "A label to attach to the low watermark line",
						},
						"watermarks": &schema.Schema{
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"value": &schema.Schema{
										Type:        schema.TypeFloat,
										Required:    true,
										Description: "Axis value where the watermark line will be displayed",
									},
									"label": &schema.Schema{
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Label to display associated with the watermark line",
									},
								},
							},
						},
					},
				},
			},
			"axis_left": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					SchemaVersion: 1,
					MigrateState:  resourceAxisMigrateState,
					Schema: map[string]*schema.Schema{
						"min_value": &schema.Schema{
							Type:        schema.TypeFloat,
							Optional:    true,
							Default:     -math.MaxFloat32,
							Description: "The minimum value for the left axis",
						},
						"max_value": &schema.Schema{
							Type:        schema.TypeFloat,
							Optional:    true,
							Default:     math.MaxFloat32,
							Description: "The maximum value for the left axis",
						},
						"label": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Label of the left axis",
						},
						"high_watermark": &schema.Schema{
							Type:        schema.TypeFloat,
							Optional:    true,
							Default:     math.MaxFloat32,
							Description: "A line to draw as a high watermark",
						},
						"high_watermark_label": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "A label to attach to the high watermark line",
						},
						"low_watermark": &schema.Schema{
							Type:        schema.TypeFloat,
							Optional:    true,
							Default:     -math.MaxFloat32,
							Description: "A line to draw as a low watermark",
						},
						"low_watermark_label": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "A label to attach to the low watermark line",
						},
						"watermarks": &schema.Schema{
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"value": &schema.Schema{
										Type:        schema.TypeFloat,
										Required:    true,
										Description: "Axis value where the watermark line will be displayed",
									},
									"label": &schema.Schema{
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Label to display associated with the watermark line",
									},
								},
							},
						},
					},
				},
			},
			"axes_precision": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     3,
				Description: "Force a specific number of significant digits in the y-axis",
			},
			"axes_include_zero": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Force y-axes to always show zero",
			},
			"on_chart_legend_dimension": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Dimension to show in the on-chart legend. On-chart legend is off unless a dimension is specified. Allowed: 'metric', 'plot_label' and any dimension.",
			},
			"legend_fields_to_hide": &schema.Schema{
				Type:          schema.TypeSet,
				Optional:      true,
				Deprecated:    "Please use legend_options_fields",
				ConflictsWith: []string{"legend_options_fields"},
				Elem:          &schema.Schema{Type: schema.TypeString},
				Description:   "List of properties that shouldn't be displayed in the chart legend (i.e. dimension names)",
			},
			"legend_options_fields": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"property": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of a property to hide or show in the data table.",
						},
						"enabled": &schema.Schema{
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "(true by default) Determines if this property is displayed in the data table.",
						},
					},
				},
				Optional:      true,
				ConflictsWith: []string{"legend_fields_to_hide"},
				Description:   "List of property and enabled flags to control the order and presence of datatable labels in a chart.",
			},
			"show_event_lines": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "(false by default) Whether vertical highlight lines should be drawn in the visualizations at times when events occurred",
			},
			"show_data_markers": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "(false by default) Show markers (circles) for each datapoint used to draw line or area charts",
			},
			"stacked": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "(false by default) Whether area and bar charts in the visualization should be stacked",
			},
			"tags": &schema.Schema{
				Type:        schema.TypeList,
				Deprecated:  "signalfx_time_chart.tags is being removed in the next release",
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Tags associated with the chart",
			},
			"plot_type": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "LineChart",
				Description:  "(LineChart by default) The default plot display style for the visualization. Must be \"LineChart\", \"AreaChart\", \"ColumnChart\", or \"Histogram\"",
				ValidateFunc: validatePlotTypeTimeChart,
			},
			"histogram_options": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Options specific to Histogram charts",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"color_theme": &schema.Schema{
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Base color theme to use for the graph.",
							ValidateFunc: validateFullPaletteColors,
						},
					},
				},
			},
			"viz_options": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Plot-level customization options, associated with a publish statement",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"label": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "The label used in the publish statement that displays the plot (metric time series data) you want to customize",
						},
						"color": &schema.Schema{
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Color to use",
							ValidateFunc: validatePerSignalColor,
						},
						"axis": &schema.Schema{
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateAxisTimeChart,
							Description:  "The Y-axis associated with values for this plot. Must be either \"right\" or \"left\"",
						},
						"plot_type": &schema.Schema{
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validatePlotTypeTimeChart,
							Description:  "(Chart plot_type by default) The visualization style to use. Must be \"LineChart\", \"AreaChart\", \"ColumnChart\", or \"Histogram\"",
						},
						"value_unit": &schema.Schema{
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateUnitTimeChart,
							Description:  "A unit to attach to this plot. Units support automatic scaling (eg thousands of bytes will be displayed as kilobytes)",
						},
						"value_prefix": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "An arbitrary prefix to display with the value of this plot",
						},
						"value_suffix": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "An arbitrary suffix to display with the value of this plot",
						},
					},
				},
			},
			"url": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL of the chart",
			},
		},

		Create: timechartCreate,
		Read:   timechartRead,
		Update: timechartUpdate,
		Delete: timechartDelete,
	}
}

/*
  Use Resource object to construct json payload in order to create a time chart
*/
func getPayloadTimeChart(d *schema.ResourceData) *chart.CreateUpdateChartRequest {
	var tags []string
	if val, ok := d.GetOk("tags"); ok {
		tags := []string{}
		for _, tag := range val.([]interface{}) {
			tags = append(tags, tag.(string))
		}
	}

	payload := &chart.CreateUpdateChartRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		ProgramText: d.Get("program_text").(string),
		Tags:        tags,
	}

	viz := getTimeChartOptions(d)
	if axesOptions := getAxesOptions(d); len(axesOptions) > 0 {
		viz.Axes = axesOptions
	}
	// There are two ways to maniplate the legend. The first is keyed from
	// `legend_fields_to_hide`. Anything in this is marked as hidden. Unspecified
	// fields default to showing up in SFx's UI.
	if legendOptions := getLegendOptions(d); legendOptions != nil {
		viz.LegendOptions = legendOptions
		// Alternatively, the `legend_options_fields` provides finer control,
		// allowing ordering and on/off toggles. This is preferred, but we keep
		// `legend_fields_to_hide` for convenience.
	} else if legendOptions := getLegendFieldOptions(d); legendOptions != nil {
		viz.LegendOptions = legendOptions
	}

	if vizOptions := getPerSignalVizOptions(d); len(vizOptions) > 0 {
		viz.PublishLabelOptions = vizOptions
	}
	if onChartLegendDim, ok := d.GetOk("on_chart_legend_dimension"); ok {
		if onChartLegendDim == "metric" {
			onChartLegendDim = "sf_originatingMetric"
		} else if onChartLegendDim == "plot_label" {
			onChartLegendDim = "sf_metric"
		}
		viz.OnChartLegendOptions = &chart.LegendOptions{
			ShowLegend:        true,
			DimensionInLegend: onChartLegendDim.(string),
		}
	}
	payload.Options = viz

	return payload
}

func getPerSignalVizOptions(d *schema.ResourceData) []*chart.PublishLabelOptions {
	viz := d.Get("viz_options").(*schema.Set).List()
	vizList := make([]*chart.PublishLabelOptions, len(viz))
	for i, v := range viz {
		v := v.(map[string]interface{})
		item := &chart.PublishLabelOptions{
			Label: v["label"].(string),
		}
		if val, ok := v["color"].(string); ok {
			if elem, ok := PaletteColors[val]; ok {
				item.PaletteIndex = int32(elem)
			}
		}
		if val, ok := v["plot_type"].(string); ok && val != "" {
			item.PlotType = val
		}
		if val, ok := v["axis"].(string); ok && val != "" {
			if val == "right" {
				item.YAxis = int32(1)
			} else {
				item.YAxis = int32(0)
			}
		}
		if val, ok := v["value_unit"].(string); ok && val != "" {
			item.ValueUnit = val
		}
		if val, ok := v["value_suffix"].(string); ok && val != "" {
			item.ValueSuffix = val
		}
		if val, ok := v["value_prefix"].(string); ok && val != "" {
			item.ValuePrefix = val
		}

		vizList[i] = item
	}
	return vizList
}

func getAxesOptions(d *schema.ResourceData) []*chart.Axes {
	axesListopts := make([]*chart.Axes, 2)
	if tfAxisOpts, ok := d.GetOk("axis_right"); ok {
		tfRightAxisOpts := tfAxisOpts.(*schema.Set).List()[0]
		tfOpt := tfRightAxisOpts.(map[string]interface{})
		axesListopts[1] = getSingleAxisOptions(tfOpt)
	} else {
		axesListopts[1] = &chart.Axes{}
	}
	if tfAxisOpts, ok := d.GetOk("axis_left"); ok {
		tfLeftAxisOpts := tfAxisOpts.(*schema.Set).List()[0]
		tfOpt := tfLeftAxisOpts.(map[string]interface{})
		axesListopts[0] = getSingleAxisOptions(tfOpt)
	} else {
		axesListopts[0] = &chart.Axes{}
	}
	return axesListopts
}

func getSingleAxisOptions(axisOpt map[string]interface{}) *chart.Axes {
	var axis *chart.Axes

	if val, ok := axisOpt["min_value"]; ok {
		if val.(float64) != -math.MaxFloat64 {
			if axis == nil {
				axis = &chart.Axes{}
			}
			axis.Min = float32(val.(float64))
		}
	}
	if val, ok := axisOpt["max_value"]; ok {
		if val.(float64) != math.MaxFloat64 {
			if axis == nil {
				axis = &chart.Axes{}
			}
			axis.Max = float32(val.(float64))
		}
	}
	if val, ok := axisOpt["label"]; ok {
		if axis == nil {
			axis = &chart.Axes{}
		}
		axis.Label = val.(string)
	}
	if val, ok := axisOpt["high_watermark"]; ok {
		if axis == nil {
			axis = &chart.Axes{}
		}
		if val.(float64) != math.MaxFloat64 {
			if axis == nil {
				axis = &chart.Axes{}
			}
			axis.HighWatermark = float32(val.(float64))
		}
	}
	if val, ok := axisOpt["high_watermark_label"]; ok {
		if axis == nil {
			axis = &chart.Axes{}
		}
		axis.HighWatermarkLabel = val.(string)
	}
	if val, ok := axisOpt["low_watermark"]; ok {
		if axis == nil {
			axis = &chart.Axes{}
		}
		if val.(float64) != -math.MaxFloat64 {
			if axis == nil {
				axis = &chart.Axes{}
			}
			axis.LowWatermark = float32(val.(float64))
		}
	}
	if val, ok := axisOpt["low_watermark_label"]; ok {
		if axis == nil {
			axis = &chart.Axes{}
		}
		axis.LowWatermarkLabel = val.(string)
	}

	return axis
}

func getTimeChartOptions(d *schema.ResourceData) *chart.Options {
	options := &chart.Options{
		Stacked: d.Get("stacked").(bool),
		Type:    "TimeSeriesChart",
	}
	if val, ok := d.GetOk("unit_prefix"); ok {
		options.UnitPrefix = val.(string)
	}
	if val, ok := d.GetOk("color_by"); ok {
		options.ColorBy = val.(string)
	}
	if val, ok := d.GetOk("show_event_lines"); ok {
		options.ShowEventLines = val.(bool)
	}
	if val, ok := d.GetOk("plot_type"); ok {
		options.DefaultPlotType = val.(string)
	}

	if val, ok := d.GetOk("axes_precision"); ok {
		options.AxisPrecision = int32(val.(int))
	}
	if val, ok := d.GetOk("axes_include_zero"); ok {
		options.IncludeZero = val.(bool)
	}

	var programOptions *chart.GeneralOptions
	if val, ok := d.GetOk("minimum_resolution"); ok {
		if programOptions == nil {
			programOptions = &chart.GeneralOptions{}
		}
		programOptions.MinimumResolution = int32(val.(int) * 1000)
	}
	if val, ok := d.GetOk("max_delay"); ok {
		if programOptions == nil {
			programOptions = &chart.GeneralOptions{}
		}
		programOptions.MaxDelay = int32(val.(int) * 1000)
	}
	if val, ok := d.GetOk("disable_sampling"); ok {
		if programOptions == nil {
			programOptions = &chart.GeneralOptions{}
		}
		programOptions.DisableSampling = val.(bool)
	}
	options.ProgramOptions = programOptions

	var timeOptions *chart.TimeDisplayOptions
	if val, ok := d.GetOk("time_range"); ok {
		timeOptions = &chart.TimeDisplayOptions{
			Range: int64(val.(int) * 1000),
			Type:  "relative",
		}
	}
	if val, ok := d.GetOk("start_time"); ok {
		timeOptions = &chart.TimeDisplayOptions{
			Start: int64(val.(int) * 1000),
			Type:  "absolute",
		}
		if val, ok := d.GetOk("end_time"); ok {
			timeOptions.End = int64(val.(int) * 1000)
		}
	}
	options.Time = timeOptions

	// dataMarkersOption := make(map[string]interface{})
	showDataMarkers := d.Get("show_data_markers").(bool)
	if chartType, ok := d.GetOk("plot_type"); ok {
		chartType := chartType.(string)
		switch chartType {
		case "AreaChart":
			options.AreaChartOptions = &chart.AreaChartOptions{
				ShowDataMarkers: showDataMarkers,
			}
		case "Histogram":
			if histogramOptions, ok := d.GetOk("histogram_options"); ok {
				hOptions := histogramOptions.(*schema.Set).List()[0].(map[string]interface{})
				if colorTheme, ok := hOptions["color_theme"].(string); ok {
					if elem, ok := FullPaletteColors[colorTheme]; ok {
						options.HistogramChartOptions = &chart.HistogramChartOptions{
							ColorThemeIndex: int32(elem),
						}
					}
				}
			}
		// Not we don't have an option for LineChart as it is the same as
		// this default
		default:
			options.LineChartOptions = &chart.LineChartOptions{
				ShowDataMarkers: showDataMarkers,
			}
		}
	}

	return options
}

func timechartCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalfxConfig)
	payload := getPayloadTimeChart(d)

	debugOutput, _ := json.Marshal(payload)
	log.Printf("[DEBUG] SignalFx: Create Time Chart Payload: %s", string(debugOutput))

	c, err := config.Client.CreateChart(payload)
	if err != nil {
		return err
	}
	// Since things worked, set the URL and move on
	appURL, err := buildAppURL(config.CustomAppURL, CHART_APP_PATH+c.Id)
	if err != nil {
		return err
	}
	d.Set("url", appURL)
	if err := d.Set("url", appURL); err != nil {
		return err
	}
	d.SetId(c.Id)

	return timechartAPIToTF(d, c)
}

func timechartRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalfxConfig)

	c, err := config.Client.GetChart(d.Id())
	if err != nil {
		return err
	}
	return timechartAPIToTF(d, c)
}

func timechartAPIToTF(d *schema.ResourceData, c *chart.Chart) error {
	debugOutput, _ := json.Marshal(c)
	log.Printf("[DEBUG] SignalFx: Got Time Chart to enState: %s", string(debugOutput))

	if err := d.Set("name", c.Name); err != nil {
		return err
	}
	if err := d.Set("description", c.Description); err != nil {
		return err
	}
	if err := d.Set("program_text", c.ProgramText); err != nil {
		return err
	}
	if err := d.Set("tags", c.Tags); err != nil {
		return err
	}
	options := c.Options

	if err := d.Set("axes_include_zero", options.IncludeZero); err != nil {
		return err
	}
	if err := d.Set("color_by", options.ColorBy); err != nil {
		return err
	}
	if err := d.Set("plot_type", options.DefaultPlotType); err != nil {
		return err
	}
	if err := d.Set("show_event_lines", options.ShowEventLines); err != nil {
		return err
	}
	if err := d.Set("stacked", options.Stacked); err != nil {
		return err
	}
	if err := d.Set("unit_prefix", options.UnitPrefix); err != nil {
		return err
	}

	if options.AreaChartOptions != nil {
		if err := d.Set("show_data_markers", options.AreaChartOptions.ShowDataMarkers); err != nil {
			return err
		}
	}
	if options.LineChartOptions != nil {
		if err := d.Set("show_data_markers", options.LineChartOptions.ShowDataMarkers); err != nil {
			return err
		}
	}
	if options.HistogramChartOptions != nil {
		color, err := getNameFromPaletteColorsByIndex(int(options.HistogramChartOptions.ColorThemeIndex))
		if err != nil {
			return err
		}
		histOptions := map[string]interface{}{
			"color_theme": color,
		}
		if err := d.Set("histogram_options", histOptions); err != nil {
			return err
		}
	}

	if len(options.Axes) > 0 {
		axisLeft := options.Axes[0]
		// We need to verify that there are real axes and not just nil
		// or zeroed structs, so we do comparison before setting each.
		if (axisLeft == nil || *axisLeft == chart.Axes{}) {
			log.Printf("[DEBUG] SignalFx: Axis Right is nil or zero, skipping")
		} else {
			if err := d.Set("axis_left", axisToMap(axisLeft)); err != nil {
				return err
			}
		}
		axisRight := options.Axes[1]
		if (axisRight == nil || *axisRight == chart.Axes{}) {
			log.Printf("[DEBUG] SignalFx: Axis Right is nil or zero, skipping")
		} else {
			log.Printf("[DEBUG] SignalFx: Axis Right is real: %v", axisRight)
			if err := d.Set("axis_right", axisToMap(axisRight)); err != nil {
				return err
			}
		}
	}

	if options.ProgramOptions != nil {
		if err := d.Set("minimum_resolution", options.ProgramOptions.MinimumResolution/1000); err != nil {
			return err
		}
		if err := d.Set("max_delay", options.ProgramOptions.MaxDelay/1000); err != nil {
			return err
		}
		if err := d.Set("disable_sampling", options.ProgramOptions.DisableSampling); err != nil {
			return err
		}
	}

	if options.Time != nil {
		if options.Time.Type == "relative" {
			if err := d.Set("time_range", options.Time.Range/1000); err != nil {
				return err
			}
		} else {
			if err := d.Set("start_time", options.Time.Start/1000); err != nil {
				return err
			}
			if err := d.Set("end_time", options.Time.End/1000); err != nil {
				return err
			}
		}
	}

	if len(options.PublishLabelOptions) > 0 {
		plos := make([]map[string]interface{}, len(options.PublishLabelOptions))
		for i, plo := range options.PublishLabelOptions {
			no, err := publishLabelOptionsToMap(plo)
			if err != nil {
				return err
			}
			plos[i] = no
		}
		if err := d.Set("viz_options", plos); err != nil {
			return err
		}
	}

	if options.LegendOptions != nil && len(options.LegendOptions.Fields) > 0 {
		fields := make([]map[string]interface{}, len(options.LegendOptions.Fields))
		for i, lo := range options.LegendOptions.Fields {
			fields[i] = map[string]interface{}{
				"property": lo.Property,
				"enabled":  lo.Enabled,
			}
		}
		if err := d.Set("legend_options_fields", fields); err != nil {
			return err
		}
	}

	if options.OnChartLegendOptions != nil {
		if err := d.Set("on_chart_legend_dimension", options.OnChartLegendOptions.DimensionInLegend); err != nil {
			return err
		}
	}

	return nil
}

func axisToMap(axis *chart.Axes) []*map[string]interface{} {
	if axis != nil {
		return []*map[string]interface{}{
			&map[string]interface{}{
				"high_watermark":       axis.HighWatermark,
				"high_watermark_label": axis.HighWatermarkLabel,
				"label":                axis.Label,
				"low_watermark":        axis.LowWatermark,
				"low_watermark_label":  axis.LowWatermarkLabel,
				"max_value":            axis.Max,
				"min_value":            axis.Min,
			},
		}
	}
	return nil
}

func publishLabelOptionsToMap(options *chart.PublishLabelOptions) (map[string]interface{}, error) {
	color, err := getNameFromPaletteColorsByIndex(int(options.PaletteIndex))
	if err != nil {
		return map[string]interface{}{}, err
	}
	axis := "left"
	if options.YAxis == 1 {
		axis = "right"
	}

	return map[string]interface{}{
		"label":        options.Label,
		"color":        color,
		"axis":         axis,
		"plot_type":    options.PlotType,
		"value_unit":   options.ValueUnit,
		"value_suffix": options.ValueSuffix,
		"value_prefix": options.ValuePrefix,
	}, nil
}

func timechartUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalfxConfig)
	payload := getPayloadTimeChart(d)

	c, err := config.Client.UpdateChart(d.Id(), payload)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] SignalFx: Update Time Chart Response: %v", c)

	// Since things worked, set the URL and move on
	appURL, err := buildAppURL(config.CustomAppURL, CHART_APP_PATH+c.Id)
	if err != nil {
		return err
	}
	if err := d.Set("url", appURL); err != nil {
		return err
	}
	d.SetId(c.Id)
	return timechartAPIToTF(d, c)
}

func timechartDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalfxConfig)

	return config.Client.DeleteChart(d.Id())
}

/*
  Validates the plot_type field against a list of allowed words.
*/
func validatePlotTypeTimeChart(v interface{}, k string) (we []string, errors []error) {
	value := v.(string)
	if value != "LineChart" && value != "AreaChart" && value != "ColumnChart" && value != "Histogram" {
		errors = append(errors, fmt.Errorf("%s not allowed; Must be \"LineChart\", \"AreaChart\", \"ColumnChart\", or \"Histogram\"", value))
	}
	return
}

/*
  Validates the axis right or left.
*/
func validateAxisTimeChart(v interface{}, k string) (we []string, errors []error) {
	value := v.(string)
	if value != "right" && value != "left" {
		errors = append(errors, fmt.Errorf("%s not allowed; must be either right or left", value))
	}
	return
}

func validateUnitTimeChart(v interface{}, k string) (we []string, errors []error) {
	value := v.(string)
	allowedWords := []string{
		"Bit",
		"Kilobit",
		"Megabit",
		"Gigabit",
		"Terabit",
		"Petabit",
		"Exabit",
		"Zettabit",
		"Yottabit",
		"Byte",
		"Kibibyte",
		"Mebibyte",
		"Gigibyte",
		"Tebibyte",
		"Pebibyte",
		"Exbibyte",
		"Zebibyte",
		"Yobibyte",
		"Nanosecond",
		"Microsecond",
		"Millisecond",
		"Second",
		"Minute",
		"Hour",
		"Day",
		"Week",
	}
	for _, word := range allowedWords {
		if value == word {
			return
		}
	}
	errors = append(errors, fmt.Errorf("%s not allowed; must be one of: %s", value, strings.Join(allowedWords, ", ")))
	return
}