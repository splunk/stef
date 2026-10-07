package tests;

import com.example.otelstef.AnyValue;
import com.example.otelstef.Exemplar;
import com.example.otelstef.ExemplarValue;
import com.example.otelstef.LogRecord;
import com.example.otelstef.Logs;
import com.example.otelstef.Point;
import com.example.otelstef.PointValue;
import net.stef.StringValue;
import org.junit.jupiter.api.Test;

import java.lang.reflect.Field;
import java.util.HashSet;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

class CloneTest {
    @Test
    void logsCloneTracksExistingAttributeValueModifications() {
        Logs source = new Logs();
        AnyValue attributeValue = new LogRecord().getBody();
        attributeValue.setInt64(1);
        source.getLog().getAttributes().append(new StringValue("key"), attributeValue);

        Logs copy = source.clone();

        assertFalse(copy.isLogModified());
        assertFalse(copy.getLog().isAttributesModified());

        elementValue(copy.getLog().getAttributes().at(0)).setInt64(2);

        assertTrue(copy.isLogModified());
        assertTrue(copy.getLog().isAttributesModified());
    }

    @Test
    void logsCloneTracksExistingNestedBodyMapValueModifications() {
        Logs source = new Logs();
        source.getLog().getBody().setType(AnyValue.Type.TypeKVList);
        AnyValue mapValue = new LogRecord().getBody();
        mapValue.setInt64(1);
        source.getLog().getBody().getKVList().append(new StringValue("key"), mapValue);

        Logs copy = source.clone();

        assertFalse(copy.isLogModified());
        assertFalse(copy.getLog().isBodyModified());

        elementValue(copy.getLog().getBody().getKVList().at(0)).setInt64(2);

        assertTrue(copy.isLogModified());
        assertTrue(copy.getLog().isBodyModified());
    }

    @Test
    void logsCloneTracksNestedLogModifications() {
        Logs source = new Logs();
        source.getLog().setTimeUnixNano(1);

        Logs copy = source.clone();

        assertFalse(copy.isLogModified());
        assertFalse(copy.getLog().isTimeUnixNanoModified());

        copy.getLog().setTimeUnixNano(2);

        assertTrue(copy.isLogModified());
        assertTrue(copy.getLog().isTimeUnixNanoModified());
    }

    @Test
    void logRecordCloneTracksNestedBodyModifications() {
        LogRecord source = new LogRecord();
        source.getBody().setInt64(1);

        LogRecord copy = source.clone();

        assertFalse(copy.isBodyModified());

        copy.getBody().setInt64(2);

        assertTrue(copy.isBodyModified());
    }

    @Test
    void oneofClonePreservesEqualityAndHashCodeAfterVariantSwitch() {
        AnyValue anyValue = new LogRecord().getBody();
        anyValue.setInt64(42);
        anyValue.setFloat64(1.0);
        assertCloneHasEqualHashCode(anyValue, anyValue.clone());

        PointValue pointValue = new Point().getValue();
        pointValue.setInt64(42);
        pointValue.setFloat64(1.0);
        assertCloneHasEqualHashCode(pointValue, pointValue.clone());

        ExemplarValue exemplarValue = new Exemplar().getValue();
        exemplarValue.setInt64(42);
        exemplarValue.setFloat64(1.0);
        assertCloneHasEqualHashCode(exemplarValue, exemplarValue.clone());
    }

    private static <T> void assertCloneHasEqualHashCode(T source, T copy) {
        assertEquals(source, copy);
        assertEquals(source.hashCode(), copy.hashCode());

        HashSet<T> values = new HashSet<>();
        values.add(source);
        assertTrue(values.contains(copy));
    }

    private static AnyValue elementValue(Object elem) {
        try {
            Field valueField = elem.getClass().getDeclaredField("value");
            valueField.setAccessible(true);
            return (AnyValue) valueField.get(elem);
        } catch (ReflectiveOperationException e) {
            throw new AssertionError(e);
        }
    }
}
