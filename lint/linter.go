package lint

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
)

type Linter struct {
	allRules     map[string]Rule
	enabledRules map[string]Rule
	problems     []*Problem
}

func NewLinter(rules ...Rule) (*Linter, error) {
	l := &Linter{
		allRules:     make(map[string]Rule, len(rules)),
		enabledRules: make(map[string]Rule, len(rules)),
	}
	for _, rule := range rules {
		if _, ok := l.allRules[rule.RuleID()]; ok {
			return nil, fmt.Errorf("duplicate rules ID %s", rule.RuleID())
		}
		l.allRules[rule.RuleID()] = rule
		l.enabledRules[rule.RuleID()] = rule
	}
	return l, nil
}

func (l *Linter) ParamFunc(name, value string) error {
	switch name {
	case "enable_all":
		for _, rule := range l.allRules {
			if _, ok := l.enabledRules[rule.RuleID()]; ok {
				return fmt.Errorf("already enabled %s", rule.RuleID())
			}
			l.enabledRules[rule.RuleID()] = rule
		}
	case "disable_all":
		l.enabledRules = make(map[string]Rule, len(l.allRules))
	case "enable":
		rule, ok := l.allRules[value]
		if !ok {
			return fmt.Errorf("unknown rule %s", rule.RuleID())
		}
		if _, ok := l.enabledRules[rule.RuleID()]; ok {
			return fmt.Errorf("already enabled %s", rule.RuleID())
		}
		l.enabledRules[rule.RuleID()] = rule
	case "disable":
		if _, ok := l.enabledRules[value]; !ok {
			return fmt.Errorf("not enabled %s", value)
		}
		delete(l.enabledRules, value)
	default:
		return fmt.Errorf("unsupported option: %s", name)
	}
	return nil
}

func (l *Linter) Run(plugin *protogen.Plugin) error {
	if err := l.initializeRules(plugin); err != nil {
		return err
	}
	for _, file := range plugin.Files {
		if !file.Generate {
			continue
		}
		if err := l.runFileRules(file); err != nil {
			return err
		}
		for _, message := range getAllMessagesInFile(file) {
			if err := l.runMessageRules(message); err != nil {
				return err
			}
			for _, field := range message.Fields {
				if err := l.runFieldRules(field); err != nil {
					return err
				}
			}
		}
		for _, enum := range getAllEnumsInFile(file) {
			if err := l.runEnumRules(enum); err != nil {
				return err
			}
			for _, enumValue := range enum.Values {
				if err := l.runEnumValueRules(enumValue); err != nil {
					return err
				}
			}
		}
		for _, service := range file.Services {
			if err := l.runServiceRules(service); err != nil {
				return err
			}
			for _, method := range service.Methods {
				if err := l.runMethodRules(method); err != nil {
					return err
				}
			}
		}
	}
	if len(l.problems) > 0 {
		return &Error{
			Plugin:   plugin,
			Problems: l.problems,
		}
	}
	return nil
}

func (l *Linter) initializeRules(gen *protogen.Plugin) error {
	for _, rule := range l.enabledRules {
		initializerRule, ok := rule.(InitializerRule)
		if !ok {
			continue
		}
		if err := initializerRule.Initialize(gen); err != nil {
			return err
		}
	}
	return nil
}

func (l *Linter) runFileRules(file *protogen.File) error {
	for _, rule := range l.enabledRules {
		fileRule, ok := rule.(FileRule)
		if !ok {
			continue
		}
		problems, err := fileRule.LintFile(file)
		if err != nil {
			return err
		}
		for _, problem := range problems {
			problem.RuleID = rule.RuleID()
			if problem.Location.SourceFile == "" {
				problem.Location = protogen.Location{SourceFile: file.Desc.Path()}
			}
			l.problems = append(l.problems, problem)
		}
	}
	return nil
}

func (l *Linter) runMessageRules(message *protogen.Message) error {
	for _, rule := range l.enabledRules {
		messageRule, ok := rule.(MessageRule)
		if !ok {
			continue
		}
		problems, err := messageRule.LintMessage(message)
		if err != nil {
			return err
		}
		for _, problem := range problems {
			problem.RuleID = rule.RuleID()
			if problem.Location.SourceFile == "" {
				problem.Location = message.Location
			}
			l.problems = append(l.problems, problem)
		}
	}
	return nil
}

func (l *Linter) runFieldRules(field *protogen.Field) error {
	for _, rule := range l.enabledRules {
		fieldRule, ok := rule.(FieldRule)
		if !ok {
			continue
		}
		problems, err := fieldRule.LintField(field)
		if err != nil {
			return err
		}
		for _, problem := range problems {
			problem.RuleID = rule.RuleID()
			if problem.Location.SourceFile == "" {
				problem.Location = field.Location
			}
			l.problems = append(l.problems, problem)
		}
	}
	return nil
}

func (l *Linter) runEnumRules(enum *protogen.Enum) error {
	for _, rule := range l.enabledRules {
		enumRule, ok := rule.(EnumRule)
		if !ok {
			continue
		}
		problems, err := enumRule.LintEnum(enum)
		if err != nil {
			return err
		}
		for _, problem := range problems {
			problem.RuleID = rule.RuleID()
			if problem.Location.SourceFile == "" {
				problem.Location = enum.Location
			}
			l.problems = append(l.problems, problem)
		}
	}
	return nil
}

func (l *Linter) runEnumValueRules(enumValue *protogen.EnumValue) error {
	for _, rule := range l.enabledRules {
		enumValueRule, ok := rule.(EnumValueRule)
		if !ok {
			continue
		}
		problems, err := enumValueRule.LintEnumValue(enumValue)
		if err != nil {
			return err
		}
		for _, problem := range problems {
			problem.RuleID = rule.RuleID()
			if problem.Location.SourceFile == "" {
				problem.Location = enumValue.Location
			}
			l.problems = append(l.problems, problem)
		}
	}
	return nil
}

func (l *Linter) runServiceRules(service *protogen.Service) error {
	for _, rule := range l.enabledRules {
		serviceRule, ok := rule.(ServiceRule)
		if !ok {
			continue
		}
		problems, err := serviceRule.LintService(service)
		if err != nil {
			return err
		}
		for _, problem := range problems {
			problem.RuleID = rule.RuleID()
			if problem.Location.SourceFile == "" {
				problem.Location = service.Location
			}
			l.problems = append(l.problems, problem)
		}
	}
	return nil
}

func (l *Linter) runMethodRules(method *protogen.Method) error {
	for _, rule := range l.enabledRules {
		methodRule, ok := rule.(MethodRule)
		if !ok {
			continue
		}
		problems, err := methodRule.LintMethod(method)
		if err != nil {
			return err
		}
		for _, problem := range problems {
			problem.RuleID = rule.RuleID()
			if problem.Location.SourceFile == "" {
				problem.Location = method.Location
			}
			l.problems = append(l.problems, problem)
		}
	}
	return nil
}

func getAllMessagesInFile(file *protogen.File) []*protogen.Message {
	result := make([]*protogen.Message, 0, len(file.Messages))
	result = append(result, file.Messages...)
	for _, message := range file.Messages {
		result = append(result, getAllMessagesInMessage(message)...)
	}
	return result
}

func getAllMessagesInMessage(message *protogen.Message) []*protogen.Message {
	result := make([]*protogen.Message, 0, len(message.Messages))
	result = append(result, message.Messages...)
	for _, childMessage := range message.Messages {
		result = append(result, getAllMessagesInMessage(childMessage)...)
	}
	return result
}

func getAllEnumsInFile(file *protogen.File) []*protogen.Enum {
	result := make([]*protogen.Enum, 0, len(file.Enums))
	result = append(result, file.Enums...)
	for _, message := range getAllMessagesInFile(file) {
		result = append(result, message.Enums...)
	}
	return result
}
