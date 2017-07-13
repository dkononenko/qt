package converter

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/therecipe/qt/internal/binding/parser"
)

func CppInputParameters(function *parser.Function) string {

	var input = make([]string, len(function.Parameters))

	for i, parameter := range function.Parameters {
		input[i] = CppInput(parameter.Name, parameter.Value, function)
	}

	return strings.Join(input, ", ")
}

func CppInputParametersForSlotInvoke(function *parser.Function) string {
	if len(function.Parameters) == 0 {
		return ""
	}

	var input = make([]string, len(function.Parameters))

	for i, parameter := range function.Parameters {
		input[i] = fmt.Sprintf("Q_ARG(%v, %v)", CppInputParametersForSlotArguments(function, parameter), CppInput(parameter.Name, parameter.Value, function))

		if c, _ := function.Class(); c.Module == parser.MOC && parser.IsPackedMap(parameter.Value) && function.IsMocFunction {
			var tHash = sha1.New()
			tHash.Write([]byte(parameter.Value))
			input[i] = strings.Replace(input[i], parser.CleanValue(parameter.Value), fmt.Sprintf("type%v", hex.EncodeToString(tHash.Sum(nil)[:3])), -1)
		}
	}

	return fmt.Sprintf(", %v", strings.Join(input, ", "))
}

func CppInputParametersForSlotArguments(function *parser.Function, parameter *parser.Parameter) string {
	switch {
	case strings.Contains(parameter.Value, "*"):
		{
			if parser.IsPackedList(parameter.Value) || parser.IsPackedMap(parameter.Value) {
				return fmt.Sprintf("%v", parser.CleanValue(parameter.Value))
			}
			return fmt.Sprintf("%v*", parser.CleanValue(parameter.Value))
		}

	case isEnum(function.ClassName(), parameter.Value):
		{
			if function.Meta == parser.SLOT && function.SignalMode == "" && parser.CleanValue(parameter.Value) == "Qt::Alignment" {
				return parser.CleanValue(parameter.Value)
			}
			return cppEnum(function, parameter.Value, false)
		}

	default:
		{
			return parser.CleanValue(parameter.Value)
		}
	}
}

func CppInputParametersForSignalConnect(function *parser.Function) string {

	var input = make([]string, len(function.Parameters))

	for i, parameter := range function.Parameters {
		if isEnum(function.ClassName(), parameter.Value) {
			input[i] = cppEnum(function, parameter.Value, true)
		} else {
			input[i] = parameter.Value
		}
	}

	return strings.Join(input, ", ")
}

func CppInputParametersForCallbackHeader(function *parser.Function) string {

	var input = make([]string, len(function.Parameters))

	for i, parameter := range function.Parameters {
		if isEnum(function.ClassName(), parameter.Value) {
			input[i] = fmt.Sprintf("%v %v", cppEnum(function, parameter.Value, true), parser.CleanName(parameter.Name, parameter.Value))
		} else {
			var c, _ = function.Class()
			if parser.IsPackedMap(parameter.Value) && c.Module == parser.MOC && function.IsMocFunction {
				var tHash = sha1.New()
				tHash.Write([]byte(parameter.Value))
				input[i] = fmt.Sprintf("%v %v", fmt.Sprintf("type%v", hex.EncodeToString(tHash.Sum(nil)[:3])), parser.CleanName(parameter.Name, parameter.Value))
			} else {
				if parser.IsPackedList(parameter.Value) || parser.IsPackedMap(parameter.Value) {
					input[i] = fmt.Sprintf("%v %v", function.OgParameters[i].Value, parser.CleanName(parameter.Name, parameter.Value))
				} else {
					input[i] = fmt.Sprintf("%v %v", parameter.Value, parser.CleanName(parameter.Name, parameter.Value))
				}
			}
		}
	}

	return strings.Join(input, ", ")
}

func CppInputParametersForCallbackBody(function *parser.Function) string {
	var input = make([]string, len(function.Parameters)+1)

	if strings.Contains(strings.Split(function.Signature, ")")[1], "const") {
		input[0] = fmt.Sprintf("const_cast<void*>(static_cast<const void*>(this))")
	} else {
		input[0] = "this"
	}

	for i, parameter := range function.Parameters {
		input[i+1] = cppOutputPacked(parameter.Name, parameter.Value, function)
	}

	return strings.Join(input, ", ")
}

func CppInputParametersForCallbackBodyPrePack(function *parser.Function) string {
	var input = make([]string, 0)

	for _, parameter := range function.Parameters {
		if packed := cppOutputPack(parameter.Name, parameter.Value, function); packed != "" {
			input = append(input, packed)
		}
	}

	return strings.Join(input, "")
}

func CppRegisterMetaType(function *parser.Function) string {
	var out = make([]string, 0)

	for _, p := range function.Parameters {
		if isEnum(function.ClassName(), p.Value) {
			out = append(out, cppEnum(function, p.Value, true))
		}
	}

	if isEnum(function.ClassName(), function.Output) {
		out = append(out, cppEnum(function, function.Output, true))
	}

	//TODO: these should ideally be registered with Q_DECLARE_METATYPE instead; also relocate use into registerTypes function
	for i := len(out) - 1; i >= 0; i-- {
		switch out[i] {
		case "QAbstractAnimation::DeletionPolicy", "QTimeLine::State", "QClipboard::Mode",
			"QImageReader::ImageReaderError", "QLocalSocket::LocalSocketError", "QSslSocket::SslMode", "QNearFieldShareManager::ShareModes",
			"QAccelerometer::AccelerationMode", "QSensor::AxesOrientationMode", "QGeoAreaMonitorSource::Error", "QGeoPositionInfoSource::Error",
			"QGeoSatelliteInfoSource::Error", "QSystemTrayIcon::MessageIcon", "QLocalSocket::LocalSocketState", "QGraphicsScene::SceneLayers",
			"QSystemTrayIcon::ActivationReason", "QGraphicsBlurEffect::BlurHints", "QAbstractItemDelegate::EndEditHint", "QDockWidget::DockWidgetFeatures",
			"QAbstract3DSeries::Mesh", "QAbstract3DInputHandler::InputView", "QAbstract3DGraph::ShadowQuality",
			"QAbstract3DGraph::SelectionFlags", "QWebSocketProtocol::CloseCode", "QSqlDriver::NotificationSource", "QCamera::LockTypes",
			"QGeoRouteReply::Error", "QWebEnginePage::Feature", "QWebEnginePage::RenderProcessTerminationStatus", "QSerialPort::Directions",
			"QPrinter::Orientation", "QPrintPreviewWidget::ViewMode", "QPrintPreviewWidget::ZoomMode", "QInAppProduct::ProductType",
			"Q3DCamera::CameraPreset", "Q3DTheme::ColorStyle", "Q3DTheme::Theme", "QAbstract3DAxis::AxisOrientation", "QAbstract3DGraph::OptimizationHints",
			"QAbstract3DGraph::ElementType", "QImage::Format", "QItemModelBarDataProxy::MultiMatchBehavior", "QSurface3DSeries::DrawFlags",
			"QAbstractBarSeries::LabelsPosition", "QScatterSeries::MarkerShape", "QWebPage::MessageSource", "QWebPage::MessageLevel",
			"QWebPage::Feature", "QItemModelSurfaceDataProxy::MultiMatchBehavior", "QCategoryAxis::AxisLabelsPosition",
			"QLegend::MarkerShape", "QDesignerFormWindowInterface::Feature":
			out = append(out[:i], out[i+1:]...)
		}
	}

	for i := range out {
		out[i] = fmt.Sprintf("\tqRegisterMetaType<%v>();", out[i])
	}

	return strings.Join(out, "\n")
}
