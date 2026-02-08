import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Trash2, Plus, GripVertical } from 'lucide-react';
import { BookingQuestion, QuestionType } from '@/types/event-type';

interface QuestionEditorProps {
    questions: BookingQuestion[];
    onChange: (questions: BookingQuestion[]) => void;
}

export function QuestionEditor({ questions, onChange }: QuestionEditorProps) {
    const addQuestion = () => {
        const newQuestion: BookingQuestion = {
            id: crypto.randomUUID(),
            label: '',
            type: 'text', // Default to text, ensuring it matches QuestionType
            required: false,
            options: [],
        };
        onChange([...questions, newQuestion]);
    };

    const updateQuestion = (index: number, updates: Partial<BookingQuestion>) => {
        const newQuestions = [...questions];
        newQuestions[index] = { ...newQuestions[index], ...updates };
        onChange(newQuestions);
    };

    const removeQuestion = (index: number) => {
        onChange(questions.filter((_, i) => i !== index));
    };

    const addOption = (qIndex: number) => {
        const newQuestions = [...questions];
        const currentOptions = newQuestions[qIndex].options || [];
        newQuestions[qIndex].options = [...currentOptions, ''];
        onChange(newQuestions);
    };

    const updateOption = (qIndex: number, oIndex: number, value: string) => {
        const newQuestions = [...questions];
        if (newQuestions[qIndex].options) {
            newQuestions[qIndex].options![oIndex] = value;
            onChange(newQuestions);
        }
    };

    const removeOption = (qIndex: number, oIndex: number) => {
        const newQuestions = [...questions];
        if (newQuestions[qIndex].options) {
            newQuestions[qIndex].options = newQuestions[qIndex].options!.filter((_, i) => i !== oIndex);
            onChange(newQuestions);
        }
    };

    return (
        <div className="space-y-4">
            {questions.map((q, index) => (
                <div key={q.id} className="flex gap-4 p-4 border rounded-lg bg-gray-50 items-start group">
                    <GripVertical className="w-5 h-5 text-gray-400 mt-2 cursor-move flex-shrink-0" />

                    <div className="flex-1 space-y-4">
                        <div className="grid grid-cols-1 md:grid-cols-12 gap-4">
                            <div className="md:col-span-5">
                                <Label className="text-xs text-muted-foreground mb-1.5 block">Label</Label>
                                <Input
                                    value={q.label}
                                    onChange={(e) => updateQuestion(index, { label: e.target.value })}
                                    placeholder="e.g. What is your phone number?"
                                    className="h-9 bg-white"
                                />
                            </div>
                            <div className="md:col-span-3">
                                <Label className="text-xs text-muted-foreground mb-1.5 block">Type</Label>
                                <Select
                                    value={q.type}
                                    onValueChange={(val) => updateQuestion(index, { type: val as QuestionType })}
                                >
                                    <SelectTrigger className="h-9 bg-white">
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="text">Text</SelectItem>
                                        <SelectItem value="phone">Phone</SelectItem>
                                        <SelectItem value="select">Select</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>
                            <div className="md:col-span-3 flex items-center h-full pt-6">
                                <div className="flex items-center space-x-2">
                                    <Switch
                                        checked={q.required}
                                        onCheckedChange={(checked) => updateQuestion(index, { required: checked })}
                                        id={`required-${q.id}`}
                                    />
                                    <Label htmlFor={`required-${q.id}`} className="text-sm font-normal cursor-pointer whitespace-nowrap">Required</Label>
                                </div>
                            </div>
                            <div className="md:col-span-1 flex justify-end pt-6">
                                <Button
                                    variant="ghost"
                                    size="icon"
                                    className="h-8 w-8 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
                                    onClick={() => removeQuestion(index)}
                                >
                                    <Trash2 className="w-4 h-4" />
                                </Button>
                            </div>
                        </div>

                        {q.type === 'select' && (
                            <div className="pl-4 border-l-2 border-primary/20 space-y-3 pt-1">
                                <Label className="text-xs font-semibold text-primary">Options</Label>
                                <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                                    {q.options?.map((opt, oIndex) => (
                                        <div key={oIndex} className="flex gap-2">
                                            <Input
                                                value={opt}
                                                onChange={(e) => updateOption(index, oIndex, e.target.value)}
                                                placeholder={`Option ${oIndex + 1}`}
                                                className="h-8 bg-white text-sm"
                                            />
                                            <Button
                                                variant="ghost"
                                                size="icon"
                                                className="h-8 w-8 text-muted-foreground hover:text-destructive"
                                                onClick={() => removeOption(index, oIndex)}
                                            >
                                                <Trash2 className="w-3 h-3" />
                                            </Button>
                                        </div>
                                    ))}
                                </div>
                                <Button
                                    type="button"
                                    variant="outline"
                                    size="sm"
                                    onClick={() => addOption(index)}
                                    className="mt-2 h-8 text-xs"
                                >
                                    <Plus className="w-3 h-3 mr-2" /> Add Option
                                </Button>
                            </div>
                        )}
                    </div>
                </div>
            ))}

            <Button type="button" variant="outline" onClick={addQuestion} className="w-full">
                <Plus className="w-4 h-4 mr-2" /> Add a Question
            </Button>
        </div>
    );
}
